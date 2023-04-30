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
	classicLeaderboard  []serverapi.LeaderboardEntry
	arenaLeaderboard    []serverapi.LeaderboardEntry
	infArenaLeaderboard []serverapi.LeaderboardEntry
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
	r.Body = http.MaxBytesReader(w, r.Body, 32*kb)
	s.httpHandler.ServeHTTP(w, r)
}

func (s *apiServer) Preload() error {
	if err := s.reloadLeaderboard(&s.classicLeaderboard, "classic"); err != nil {
		return err
	}
	if err := s.reloadLeaderboard(&s.arenaLeaderboard, "arena"); err != nil {
		return err
	}
	if err := s.reloadLeaderboard(&s.infArenaLeaderboard, "inf_arena"); err != nil {
		return err
	}
	return nil
}

func (s *apiServer) InitDatabases() error {
	sqliteConnect := func(dbPath string) (*sql.DB, error) {
		conn, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			return nil, fmt.Errorf("open %q: %w", dbPath, err)
		}
		if err := conn.Ping(); err != nil {
			return nil, fmt.Errorf("ping %q: %w", dbPath, err)
		}
		return conn, nil
	}

	queueDBPath := filepath.Join(s.dataFolder, "queue.db")
	queueConn, err := sqliteConnect(queueDBPath)
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
		conn, err := sqliteConnect(dbPath)
		if err != nil {
			return err
		}
		db := &seasonDB{
			id:   i,
			conn: conn,
		}
		if err := db.PrepareQueries(); err != nil {
			return err
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
			if err := s.reloadLeaderboard(&s.classicLeaderboard, "classic"); err != nil {
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
			if err := s.reloadLeaderboard(&s.arenaLeaderboard, "arena"); err != nil {
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
			if err := s.reloadLeaderboard(&s.infArenaLeaderboard, "inf_arena"); err != nil {
				s.logger.Error("inf_arena leaderboard reload: %v", err)
				delayMultiplier += floatRange(s.rand, 0.5, 1.5)
			} else {
				s.logger.Info("reloaded inf_arena leaderboard")
			}
			untilInfArenaLeaderboardUpdate = s.intervalLeaderboardUpdate() * delayMultiplier
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
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(runsimBinaryName)
	cmd.Stdin = bytes.NewReader(uncompressedReplayData)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return false, err
	}
	// The simulation should never take that long, but better be safe than sorry.
	timer := time.AfterFunc(15*time.Second, func() {
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

	// Now we can delete the replay from the queue and add
	// verified results to the database.
	// TODO: this should be done in a transaction.
	drones := strings.Join(replayData.Config.Tier2Recipes, ",")
	err = db.UpdatePlayerScore(replayData.Config.RawGameMode, playerName, drones, result.Score, replayData.Config.DifficultyScore, result.Time)
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
	if n >= len(leaderboard) {
		n = len(leaderboard)
	}
	return leaderboard[:n]
}

func (s *apiServer) getBoardForMode(mode string) []serverapi.LeaderboardEntry {
	switch mode {
	case "classic":
		return s.classicLeaderboard
	case "arena":
		return s.arenaLeaderboard
	case "inf_arena":
		return s.infArenaLeaderboard
	}
	return nil
}

func (s *apiServer) NumBoardPlayers(mode string) int {
	return len(s.getBoardForMode(mode))
}

func (s *apiServer) PlayerBoard(mode, name string, score int) ([]serverapi.LeaderboardEntry, error) {
	entries := s.getBoardForMode(mode)
	if len(entries) == 0 {
		return nil, nil
	}

	i := sort.Search(len(entries), func(i int) bool {
		return entries[i].Score <= score
	})
	if i >= len(entries) || entries[i].Score != score {
		return nil, errNotFound
	}

	// Several players can have identical score.
	playerIndex := -1
	for i > 0 && entries[i-1].Score == score {
		i--
	}
	for j := i; j < len(entries) && entries[j].Score == score; j++ {
		if entries[j].PlayerName == name {
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
	if i == len(entries)-1 {
		from = i - 9
		if from < 0 {
			from = 0
		}
		to = len(entries)
	} else {
		from = i - 8
		to = i + 2
		if from < 0 {
			to += -from
			from = 0
		}
		if to > len(entries) {
			to = len(entries)
		}
	}

	return entries[from:to], nil
}

func (s *apiServer) reloadLeaderboard(dst *[]serverapi.LeaderboardEntry, mode string) error {
	s.leaderboardMu.Lock()
	defer s.leaderboardMu.Unlock()

	entries, err := s.getSeasonDB(currentSeason).AllScores(mode)
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

	*dst = entries
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
		data, err := json.Marshal(v)
		if err != nil {
			s.writeError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func (s *apiServer) corsAllowed(origin string) bool {
	const itchioWebGames = "https://v6p9d9t4.ssl.hwcdn.net"

	switch origin {
	case "http://localhost:8080", itchioWebGames:
		return true
	default:
		return false
	}
}
