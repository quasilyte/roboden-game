package clientkit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/httpfetch"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/session"
)

func GetLeaderboard(state *session.State, gameMode string) (*serverapi.LeaderboardResp, error) {
	var u url.URL
	u.Host = state.ServerHost
	u.Scheme = state.ServerProtocol
	u.Path = path.Join(state.ServerPath, "get-player-board")
	q := u.Query()
	q.Add("season", strconv.Itoa(gamedata.SeasonNumber))
	q.Add("mode", gameMode)
	q.Add("name", state.Persistent.PlayerName)
	u.RawQuery = q.Encode()

	data, err := httpfetch.GetBytes(u.String())
	if err != nil {
		return nil, err
	}
	var resp serverapi.LeaderboardResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func enqueueReplay(state *session.State, replay serverapi.GameReplay) {
	key := fmt.Sprintf("queued_replay_%d", state.Persistent.NumPendingSubmissions)
	state.Persistent.NumPendingSubmissions++
	state.Context.SaveGameData(key, replay)
	state.Context.SaveGameData("save", state.Persistent)
}

func SendOrEnqueueScore(state *session.State, replay serverapi.GameReplay) {
	sendResult, err := SendScore(state, replay)
	if err != nil || sendResult.TryAgain {
		if err != nil {
			state.Logf("sending game replay failed: %v", err)
		} else {
			state.Logf("the server asked to try again later")
		}
		enqueueReplay(state, replay)
		return
	}
	state.Logf("queued game replay successfully")
}

type SendScoreResult struct {
	TryAgain bool
	Queued   bool
}

func SendScore(state *session.State, replay serverapi.GameReplay) (SendScoreResult, error) {
	var u url.URL
	u.Host = state.ServerHost
	u.Scheme = state.ServerProtocol
	u.Path = path.Join(state.ServerPath, "save-player-score")
	q := u.Query()
	q.Add("season", strconv.Itoa(gamedata.SeasonNumber))
	q.Add("mode", replay.Config.RawGameMode)
	q.Add("name", state.Persistent.PlayerName)
	u.RawQuery = q.Encode()

	var result SendScoreResult

	replayData, err := json.Marshal(replay)
	if err != nil {
		return result, err
	}

	resp, err := httpfetch.PostJSON(u.String(), replayData)
	if err != nil {
		// Probably a network issue; or a server is down.
		// It's worth trying again.
		result.TryAgain = true
		return result, err
	}

	switch resp.Code {
	case http.StatusTooManyRequests:
		// Server asks to try this again.
		result.TryAgain = true
		return result, nil
	case http.StatusOK:
		var responseInfo serverapi.SavePlayerScoreResp
		if err := json.Unmarshal(resp.Data, &responseInfo); err != nil {
			return result, err
		}
		result.Queued = responseInfo.Queued
		return result, nil
	default:
		return result, nil
	}
}
