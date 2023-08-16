package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/serverapi"
)

// requestHandler implements the API by handling the requests.
// It's shared between the requests, so it shouldn't have unsynchronized state.
type requestHandler struct {
	server *apiServer

	versionResponse []byte
}

func newRequestHandler(s *apiServer) *requestHandler {
	versionResponse := []byte(fmt.Sprintf(`{"version":"%s"}`, CommitHash))

	return &requestHandler{
		server:          s,
		versionResponse: versionResponse,
	}
}

func (h *requestHandler) HandleVersion(r *http.Request) (any, error) {
	h.server.metrics.IncReqVersion()
	return h.versionResponse, nil
}

func (h *requestHandler) HandleGetPlayerBoard(r *http.Request) (any, error) {
	h.server.metrics.IncReqGetPlayerBoard()

	seasonParam := r.URL.Query().Get("season")
	if seasonParam == "" {
		return nil, errBadParams
	}
	modeParam := r.URL.Query().Get("mode")
	switch modeParam {
	case "classic", "arena", "inf_arena", "reverse":
		// OK
	default:
		return nil, errBadParams
	}
	playerName := r.URL.Query().Get("name")
	seasonNumber, err := strconv.Atoi(seasonParam)
	if err != nil {
		return nil, errBadParams
	}

	db := h.server.getSeasonDB(seasonNumber)
	if db == nil {
		return nil, errBadParams
	}

	resp := &serverapi.LeaderboardResp{
		NumSeasons: h.server.NumSeasons(),
		NumPlayers: h.server.NumBoardPlayers(modeParam),
	}
	playerName = strings.TrimSpace(playerName)
	if playerName == "" || !gamedata.IsValidUsername(playerName) {
		resp.Entries = h.server.Top10(modeParam)
		return resp, nil
	}
	playerScore := db.PlayerScore(modeParam, playerName)
	if playerScore == -1 {
		resp.Entries = h.server.Top10(modeParam)
		return resp, nil
	}
	leaderboardEntries, err := h.server.PlayerBoard(modeParam, playerName, playerScore)
	if err != nil {
		return nil, err
	}
	resp.Entries = leaderboardEntries
	return resp, nil
}

func (h *requestHandler) HandleGetBoard(r *http.Request) (any, error) {
	h.server.metrics.IncReqGetBoard()

	seasonParam := r.URL.Query().Get("season")
	if seasonParam == "" {
		return nil, errBadParams
	}
	modeParam := r.URL.Query().Get("mode")
	switch modeParam {
	case "classic", "arena", "inf_arena", "reverse":
		// OK
	default:
		return nil, errBadParams
	}
	seasonNumber, err := strconv.Atoi(seasonParam)
	if err != nil {
		return nil, errBadParams
	}
	if seasonNumber != currentSeason {
		return nil, errBadParams
	}

	board := h.server.getBoardForMode(modeParam)
	return board.json, nil
}

func (h *requestHandler) HandleSavePlayerScore(r *http.Request) (any, error) {
	h.server.metrics.IncReqSavePlayerScore()

	if r.Method != http.MethodPost {
		return nil, errBadHTTPMethod
	}

	seasonParam := r.URL.Query().Get("season")
	if seasonParam == "" {
		return nil, errBadParams
	}
	playerName := r.URL.Query().Get("name")
	playerName = strings.TrimSpace(playerName)
	if playerName == "" || !gamedata.IsValidUsername(playerName) {
		return nil, errBadParams
	}
	seasonNumber, err := strconv.Atoi(seasonParam)
	if err != nil {
		return nil, errBadParams
	}
	if seasonNumber != currentSeason {
		return nil, errBadParams
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errBadParams
	}

	var gameReplay serverapi.GameReplay
	if err := json.Unmarshal(data, &gameReplay); err != nil {
		return nil, errBadParams
	}
	if err := h.isValidGameReplay(gameReplay); err != nil {
		return nil, err
	}

	// TODO: hashcash check here to deal with spammers?

	// Check if we have the right runsim binary for this match.
	runsimBinaryName := filepath.Join(h.server.runsimFolder, fmt.Sprintf("runsim_%d", gameReplay.GameVersion))
	if !fileExists(runsimBinaryName) {
		h.server.logger.Info("unsupported game build %v is requested", gameReplay.GameVersion)
		return nil, errUnsupportedBuild
	}

	db := h.server.getSeasonDB(seasonNumber)

	// See if there is any free space left in our queue.
	queueSize, err := h.server.queue.Count()
	if err != nil {
		return nil, err
	}
	if queueSize > 512 {
		h.server.metrics.IncNumReplaysRejected()
		h.server.logger.Info("rejected %q replay, the queue is full", playerName)
		return nil, errQueueIsFull
	}

	replayChecksum := h.calcReplayChecksum(&gameReplay)
	checksumOwner, err := h.server.queue.ChecksumOwner(replayChecksum)
	if err != nil {
		return nil, err
	}
	if checksumOwner != "" {
		if checksumOwner != playerName {
			h.server.logger.Info("got duplicate checksum from %q, current owner is %q", playerName, checksumOwner)
			return nil, errBadParams
		}
		h.server.logger.Info("%q sent existing replay", playerName)
		// Make the client happy. Tell them that we're working on this replay.
		resp := &serverapi.SavePlayerScoreResp{Queued: true}
		return resp, nil
	}

	resp := &serverapi.SavePlayerScoreResp{}

	// Now check if it actually makes sense to calculate the score.
	// If claimed score is less than the current record for the player,
	// don't bother calculating this submission.
	playerScore := db.PlayerScore(gameReplay.Config.RawGameMode, playerName)
	resp.CurrentHighscore = playerScore
	if playerScore > gameReplay.Results.Score {
		// Not queued, but the current score is better than submitted result.
		// The client should figure things out.
		return resp, nil
	}

	// Don't allow more than a few simulations be enqueued for a single player.
	// This way, we try to decrease the spam effect and keep the queue
	// available for more players.
	countForPlayer, err := h.server.queue.CountForPlayer(playerName)
	if err != nil {
		return nil, err
	}
	if countForPlayer >= 3 {
		h.server.logger.Info("player %q sends too many replays", playerName)
		return nil, errQueueIsFull
	}

	// If everything looks good so far, put it into the queue.
	// Use the compressed data we've read from the request body to avoid
	// redundant encoding/compression.
	timestamp := time.Now().Unix()
	if err := h.server.queue.PushRaw(replayChecksum, playerName, timestamp, data, false); err != nil {
		return nil, err
	}

	h.server.logger.Info("added %q replay to the queue (mode=%q score=%d)", playerName, gameReplay.Config.RawGameMode, gameReplay.Results.Score)
	h.server.metrics.IncNumReplaysQueued()
	resp.Queued = true
	return resp, err
}

func (h *requestHandler) isValidGameReplay(r serverapi.GameReplay) error {
	// This is just a superficial check before putting this replay
	// into the queue. The game simulator will apply proper validation.

	if !gamedata.IsValidReplay(r) {
		return errBadParams
	}
	if !gamedata.IsSendableReplay(r) {
		return errBadParams
	}

	return nil
}

func (h *requestHandler) calcReplayChecksum(replay *serverapi.GameReplay) string {
	buf := make([]byte, 0, 256)

	buf = strconv.AppendInt(buf, replay.Config.Seed, 10)
	buf = append(buf, '/')
	buf = append(buf, replay.Config.RawGameMode...)
	buf = append(buf, '/')
	for _, drone := range replay.Config.Tier2Recipes {
		buf = append(buf, drone...)
		buf = append(buf, '$')
	}
	buf = append(buf, replay.Config.TurretDesign...)
	buf = append(buf, '$')
	buf = strconv.AppendInt(buf, int64(len(replay.Actions)), 10)
	buf = append(buf, '@')
	buf = strconv.AppendInt(buf, int64(replay.Results.Score), 10)
	buf = append(buf, ':')
	buf = strconv.AppendInt(buf, int64(replay.Results.Ticks), 10)
	buf = append(buf, ':')
	buf = strconv.AppendInt(buf, int64(replay.Results.Time), 10)
	buf = append(buf, ';')

	return sha1encode(buf)
}
