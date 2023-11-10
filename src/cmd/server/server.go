package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/quasilyte/roboden-game/serverapi"
	"github.com/quasilyte/roboden-game/sqliteutil"
)

// apiServer is something that lives detached from the HTTP request serving
// routine. It updates in-memory caches periodically, etc.
// API request handler uses the server to get an access to this cached data.
// The access is synchronized (when necessary).
type apiServer struct {
	queue *replayQueue

	httpHandler http.Handler

	seasons    []*seasonDB
	dataFolder string
	logger     logger

	sleepStart time.Time
	stop       int64

	runsimFolder string

	rand *rand.Rand

	metricsFile string
	metrics     *serverMetrics

	leaderboardMu       sync.RWMutex
	classicLeaderboard  *leaderboardData
	arenaLeaderboard    *leaderboardData
	infArenaLeaderboard *leaderboardData
	reverseLeaderboard  *leaderboardData
}

type leaderboardData struct {
	mode    string
	entries []serverapi.LeaderboardEntry
	json    []byte
}

type serverConfig struct {
	httpHandler  http.Handler
	runsimFolder string
	dataFolder   string
	metricsFile  string
	logger       logger
}

func newAPIServer(config serverConfig) *apiServer {
	s := &apiServer{
		httpHandler:  config.httpHandler,
		dataFolder:   config.dataFolder,
		runsimFolder: config.runsimFolder,
		logger:       config.logger,
		rand:         rand.New(rand.NewSource(time.Now().Unix())),
		metrics:      &serverMetrics{},
		metricsFile:  config.metricsFile,

		classicLeaderboard:  &leaderboardData{mode: "classic"},
		arenaLeaderboard:    &leaderboardData{mode: "arena"},
		infArenaLeaderboard: &leaderboardData{mode: "inf_arena"},
		reverseLeaderboard:  &leaderboardData{mode: "reverse"},
	}
	return s
}

func (s *apiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		if origin := r.Header.Get("Origin"); s.corsAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, HEAD")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Origin, X-Requested-With, Content-Type, Accept")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	const kb = 1024
	r.Body = http.MaxBytesReader(w, r.Body, 128*kb)
	s.httpHandler.ServeHTTP(w, r)
}

func (s *apiServer) Preload() error {
	if err := s.reloadLeaderboard(s.classicLeaderboard); err != nil {
		return err
	}
	if err := s.reloadLeaderboard(s.arenaLeaderboard); err != nil {
		return err
	}
	if err := s.reloadLeaderboard(s.infArenaLeaderboard); err != nil {
		return err
	}
	if err := s.reloadLeaderboard(s.reverseLeaderboard); err != nil {
		return err
	}
	return nil
}

func (s *apiServer) InitDatabases() error {
	queueDBPath := filepath.Join(s.dataFolder, "queue.db")
	queueConn, err := sqliteutil.Connect(queueDBPath)
	if err != nil {
		return err
	}
	s.queue = newReplayQueue(queueConn)
	if err := s.queue.PrepareQueries(); err != nil {
		return fmt.Errorf("prepare queue queries: %w", err)
	}

	for i := 0; i <= currentSeason; i++ {
		dbFilename := fmt.Sprintf("season%d.db", i)
		dbPath := filepath.Join(s.dataFolder, dbFilename)
		conn, err := sqliteutil.Connect(dbPath)
		if err != nil {
			return fmt.Errorf("season%d: %w", i, err)
		}
		db := &seasonDB{
			id:   i,
			conn: conn,
		}
		if err := db.PrepareQueries(); err != nil {
			return fmt.Errorf("season%d: %w", i, err)
		}
		s.seasons = append(s.seasons, db)
		s.logger.Info("opened %s", dbFilename)
	}
	return nil
}

func (s *apiServer) intervalMetricsFlush() float64 {
	return floatRange(s.rand, 2*60, 6*60)
}

func (s *apiServer) intervalLeaderboardUpdate() float64 {
	return floatRange(s.rand, 45, 5*60)
}

func (s *apiServer) intervalLogRotate() float64 {
	return floatRange(s.rand, 20, 40)
}

func (s *apiServer) intervalRunReplay() float64 {
	return floatRange(s.rand, 5, 10)
}

func (s *apiServer) Stop() {
	atomic.StoreInt64(&s.stop, 1)
}

func (s *apiServer) BackgroundTask() {
	untilClassicLeaderboardUpdate := s.intervalLeaderboardUpdate()
	untilArenaLeaderboardUpdate := s.intervalLeaderboardUpdate()
	untilInfArenaLeaderboardUpdate := s.intervalLeaderboardUpdate()
	untilReverseLeaderboardUpdate := s.intervalLeaderboardUpdate()
	untilMetricsFlush := s.intervalMetricsFlush()
	untilLogRotate := s.intervalLogRotate()
	untilRunReplay := s.intervalRunReplay()

	for {
		if atomic.LoadInt64(&s.stop) != 0 {
			s.logger.Info("stopping the server")
			return
		}
		// Sleet with a random jitter.
		secondsToSleep := 3.0 * (s.rand.Float64() + 0.4)
		s.sleepStart = time.Now()
		time.Sleep(time.Second * time.Duration(secondsToSleep))
		timeSlept := time.Since(s.sleepStart)
		secondsSlept := timeSlept.Seconds()

		s.metrics.data.Uptime += secondsSlept

		untilClassicLeaderboardUpdate -= secondsSlept
		if untilClassicLeaderboardUpdate <= 0 {
			delayMultiplier := 1.0
			if err := s.reloadLeaderboard(s.classicLeaderboard); err != nil {
				s.logger.Error("classic leaderboard reload: %v", err)
				delayMultiplier += floatRange(s.rand, 0.5, 1.5)
			} else {
				s.logger.Info("reloaded classic leaderboard")
			}
			untilClassicLeaderboardUpdate = s.intervalLeaderboardUpdate() * delayMultiplier
			continue
		}
		untilArenaLeaderboardUpdate -= secondsSlept
		if untilArenaLeaderboardUpdate <= 0 {
			delayMultiplier := 1.0
			if err := s.reloadLeaderboard(s.arenaLeaderboard); err != nil {
				s.logger.Error("arena leaderboard reload: %v", err)
				delayMultiplier += floatRange(s.rand, 0.5, 1.5)
			} else {
				s.logger.Info("reloaded arena leaderboard")
			}
			untilArenaLeaderboardUpdate = s.intervalLeaderboardUpdate() * delayMultiplier
			continue
		}
		untilInfArenaLeaderboardUpdate -= secondsSlept
		if untilInfArenaLeaderboardUpdate <= 0 {
			delayMultiplier := 1.0
			if err := s.reloadLeaderboard(s.infArenaLeaderboard); err != nil {
				s.logger.Error("inf_arena leaderboard reload: %v", err)
				delayMultiplier += floatRange(s.rand, 0.5, 1.5)
			} else {
				s.logger.Info("reloaded inf_arena leaderboard")
			}
			untilInfArenaLeaderboardUpdate = s.intervalLeaderboardUpdate() * delayMultiplier
			continue
		}
		untilReverseLeaderboardUpdate -= secondsSlept
		if untilReverseLeaderboardUpdate <= 0 {
			delayMultiplier := 1.0
			if err := s.reloadLeaderboard(s.reverseLeaderboard); err != nil {
				s.logger.Error("reverse leaderboard reload: %v", err)
				delayMultiplier += floatRange(s.rand, 0.5, 1.5)
			} else {
				s.logger.Info("reloaded reverse leaderboard")
			}
			untilReverseLeaderboardUpdate = s.intervalLeaderboardUpdate() * delayMultiplier
			continue
		}

		untilMetricsFlush -= secondsSlept
		if untilMetricsFlush <= 0 {
			delayMultiplier := 1.0
			if err := s.doMetricsFlush(); err != nil {
				s.logger.Error("metrics flush: %v", err)
				delayMultiplier += floatRange(s.rand, 1.5, 3.5)
			} else {
				s.logger.Info("metrics flushed (seq=%d)", s.metrics.data.MetricsSeq)
			}
			untilMetricsFlush = s.intervalMetricsFlush() * delayMultiplier
			continue
		}

		untilLogRotate -= secondsSlept
		if untilLogRotate <= 0 {
			rotated, err := s.doLogRotate()
			if err != nil {
				s.logger.Error("rotate log file: %v", err)
			} else if rotated {
				s.logger.Info("rotated log file")
			} else {
				s.logger.Info("no need to do a log rotate")
			}
			untilLogRotate = s.intervalLogRotate()
			continue
		}

		untilRunReplay -= secondsSlept
		if untilRunReplay <= 0 {
			delayMultiplier := 1.0
			replayed, err := s.doRunReplay()
			if err != nil {
				delayMultiplier += floatRange(s.rand, 2.5, 4)
				s.logger.Error("run replay: %v", err)
			} else if replayed {
				s.logger.Info("executed a replay")
			} else {
				delayMultiplier += floatRange(s.rand, 1, 2)
				s.logger.Info("no replays to execute")
			}
			untilRunReplay = s.intervalRunReplay() * delayMultiplier
			continue
		}
	}
}

func (s *apiServer) doRunReplay() (bool, error) {
	replayID, playerName, compressedReplayData, err := s.queue.Get()
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	uncompressedReplayData, err := gzipUncompress(compressedReplayData)
	if err != nil {
		return false, err
	}

	var replayData serverapi.GameReplay
	if err := json.Unmarshal(uncompressedReplayData, &replayData); err != nil {
		s.metrics.IncNumReplaysFailed()
		s.logger.Error("found malformed replay json data with id=%d: %v", replayID, err)
		// This should never happen, since we unmarhalled the data
		// before saving it to the queue.
		// Although if it does happen, let's remove the entry so it doesn't happen again.
		if err := s.queue.Delete(replayID, playerName); err != nil {
			s.logger.Error("can't delete malformed replay with id=%d: %v", replayID, err)
			return false, err
		}
		return false, err
	}

	seasonNumber := seasonByBuild(replayData.GameVersion)
	db := s.getSeasonDB(seasonNumber)
	if db == nil {
		s.metrics.IncNumReplaysFailed()
		archivedAt := time.Now().Unix()
		if err := s.queue.Archive(replayID, playerName, archivedAt, compressedReplayData, archiveMismatchingResults); err != nil {
			s.logger.Error("can't archive bad season replay with id=%d: %v", replayID, err)
			return false, err
		}
		s.logger.Info("archived bad season (%d) replay with id=%d", seasonNumber, replayID)
		return true, nil
	}

	// See whether we have a runner for this replay.
	// The server should check this beforehand, but bad things can happen:
	// we may not have this binary anymore.
	runsimBinaryName := filepath.Join(s.runsimFolder, fmt.Sprintf("runsim_%d", replayData.GameVersion))
	if !fileExists(runsimBinaryName) {
		s.metrics.IncNumReplaysFailed()
		archivedAt := time.Now().Unix()
		if err := s.queue.Archive(replayID, playerName, archivedAt, compressedReplayData, archiveUnsupportedBuild); err != nil {
			s.logger.Error("can't archive unsupported build replay with id=%d: %v", replayID, err)
			return false, err
		}
		s.logger.Info("archived unsupported build (%d) replay with id=%d", replayData.GameVersion, replayID)
		return false, nil
	}

	start := time.Now()
	timeout := 30 * time.Second
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var runsimArgs []string
	if replayData.Config.RawGameMode == "inf_arena" {
		// Infinite arenas may take much longer to simulate due to
		// their "almost infinite" nature.
		runsimArgs = append(runsimArgs, "--timeout=60")
		timeout = 60 * time.Second
	}
	cmd := exec.Command(runsimBinaryName, runsimArgs...)
	cmd.Stdin = bytes.NewReader(uncompressedReplayData)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return false, err
	}
	// The simulation should never take that long, but better be safe than sorry.
	timer := time.AfterFunc(timeout, func() {
		cmd.Process.Kill()
	})
	err = cmd.Wait()
	timer.Stop()
	elapsed := time.Since(start)
	if err != nil {
		s.metrics.IncNumReplaysFailed()
		archivedAt := time.Now().Unix()
		if err := s.queue.Archive(replayID, playerName, archivedAt, compressedReplayData, archiveExecError); err != nil {
			s.logger.Error("can't archive bad-exec replay with id=%d: %v", replayID, err)
			return true, err
		}
		s.logger.Info("archived errored replay with id=%d", replayID)
		return true, fmt.Errorf("failed to execute runsim: %s: %w", stderr.String(), err)
	}

	s.logger.Info("simulation took %v", elapsed)
	s.metrics.IncNumReplaysCompleted()

	var result serverapi.GameResults
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return true, fmt.Errorf("unmarshal runsim results: %w", err)
	}

	if result != replayData.Results {
		s.metrics.IncNumReplaysFailed()
		archivedAt := time.Now().Unix()
		if err := s.queue.Archive(replayID, playerName, archivedAt, compressedReplayData, archiveMismatchingResults); err != nil {
			s.logger.Error("can't archive mis-simulated replay with id=%d: %v", replayID, err)
			return false, err
		}
		s.logger.Info("archived mismatching results replay with id=%d", replayID)
		return true, nil
	}

	platform := replayData.Platform
	if platform == "" {
		platform = "Steam"
	}

	difficulty := replayData.Config.DifficultyScore
	savedReplayID := 0
	if result.Score >= 1800 && difficulty >= 200 {
		// Save only interesting enough replay data.
		savedReplayID = replayID
		archivedAt := time.Now().Unix()
		if err := s.queue.GoodArchive(replayID, playerName, archivedAt, compressedReplayData); err != nil {
			// A failure to archive is not a show stopper.
			s.logger.Error("can't archive interesting replay with id=%d: %v", replayID, err)
		} else {
			s.logger.Info("archived interesting replay with id=%d", replayID)
		}
	}

	// Now we can delete the replay from the queue and add
	// verified results to the database.
	// TODO: this should be done in a transaction.
	drones := strings.Join(replayData.Config.Tier2Recipes, ",")
	err = db.UpdatePlayerScore(replayData.Config.RawGameMode, playerName, savedReplayID, drones, result.Score, difficulty, result.Time, platform)
	if err != nil {
		return true, err
	}
	if err := s.queue.Delete(replayID, playerName); err != nil {
		return true, err
	}

	return true, nil
}

func (s *apiServer) doLogRotate() (bool, error) {
	const kb = 1024
	if s.logger.GetSize() < 256*kb {
		return false, nil
	}
	return true, s.logger.Rotate()
}

func (s *apiServer) doMetricsFlush() error {
	s.metrics.data.MetricsSeq++
	jsonData, err := json.Marshal(s.metrics.data)
	if err != nil {
		return err
	}
	return os.WriteFile(s.metricsFile, jsonData, 0o666)
}

func (s *apiServer) Top10(mode string) []serverapi.LeaderboardEntry {
	leaderboard := s.getBoardForMode(mode)
	n := 10
	if n >= len(leaderboard.entries) {
		n = len(leaderboard.entries)
	}
	return leaderboard.entries[:n]
}

func (s *apiServer) getBoardForMode(mode string) *leaderboardData {
	switch mode {
	case "classic":
		return s.classicLeaderboard
	case "arena":
		return s.arenaLeaderboard
	case "inf_arena":
		return s.infArenaLeaderboard
	case "reverse":
		return s.reverseLeaderboard
	}
	return nil
}

func (s *apiServer) NumBoardPlayers(mode string) int {
	return len(s.getBoardForMode(mode).entries)
}

func (s *apiServer) PlayerBoard(mode, name string, score int) ([]serverapi.LeaderboardEntry, error) {
	board := s.getBoardForMode(mode)
	if len(board.entries) == 0 {
		return nil, nil
	}

	i := sort.Search(len(board.entries), func(i int) bool {
		return board.entries[i].Score <= score
	})
	if i >= len(board.entries) || board.entries[i].Score != score {
		return nil, errNotFound
	}

	// Several players can have identical score.
	playerIndex := -1
	for i > 0 && board.entries[i-1].Score == score {
		i--
	}
	for j := i; j < len(board.entries) && board.entries[j].Score == score; j++ {
		if board.entries[j].PlayerName == name {
			playerIndex = j
			break
		}
	}
	if playerIndex == -1 {
		return nil, errBadParams
	}
	i = playerIndex

	var from int
	var to int
	if i == len(board.entries)-1 {
		from = i - 9
		if from < 0 {
			from = 0
		}
		to = len(board.entries)
	} else {
		from = i - 8
		to = i + 2
		if from < 0 {
			to += -from
			from = 0
		}
		if to > len(board.entries) {
			to = len(board.entries)
		}
	}

	return board.entries[from:to], nil
}

func (s *apiServer) reloadLeaderboard(leaderboard *leaderboardData) error {
	s.leaderboardMu.Lock()
	defer s.leaderboardMu.Unlock()

	entries, err := s.getSeasonDB(currentSeason).AllScores(leaderboard.mode)
	if err != nil {
		return err
	}

	prevScore := 0
	rank := 0
	for i := range entries {
		e := &entries[i]
		if prevScore == 0 || prevScore > e.Score {
			rank++
			prevScore = e.Score
		}
		e.Rank = rank
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	leaderboard.json = data
	leaderboard.entries = entries

	return nil
}

func (s *apiServer) NumSeasons() int {
	return len(s.seasons)
}

func (s *apiServer) getSeasonDB(i int) *seasonDB {
	if i >= 0 && i < len(s.seasons) {
		return s.seasons[i]
	}
	return nil
}

func (s *apiServer) writeError(w http.ResponseWriter, err error) {
	s.metrics.IncNumReqErrors()

	switch err {
	case errBadParams:
		w.WriteHeader(http.StatusBadRequest)
	case errNotFound:
		w.WriteHeader(http.StatusNotFound)
	case errBadHTTPMethod:
		w.WriteHeader(http.StatusMethodNotAllowed)
	case errQueueIsFull:
		w.WriteHeader(http.StatusTooManyRequests)
	default:
		s.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *apiServer) NewHandler(f func(*http.Request) (any, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		v, err := f(r)
		if origin := r.Header.Get("origin"); s.corsAllowed(origin) {
			w.Header().Set("Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		if err != nil {
			s.writeError(w, err)
			return
		}
		if v == nil {
			w.Header().Set("Content-Type", "application/json")
			return
		}

		var data []byte
		if vData, ok := v.([]byte); ok {
			data = vData
		} else {
			data, err = json.Marshal(v)
			if err != nil {
				s.writeError(w, err)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func (s *apiServer) corsAllowed(origin string) bool {
	// If it stops working, try itch.zone
	// See https://itch.io/t/2928588/cors-setting-for-a-webgl-game
	const itchioWebGames = "https://v6p9d9t4.ssl.hwcdn.net"

	switch origin {
	case "https://roboden-game.github.io":
		// Used for the online leaderboard.
		return true

	case "http://localhost:8080", "http://localhost:8000":
		// The usual local debug servers.
		return true

	case itchioWebGames:
		return true

	default:
		return false
	}
}
