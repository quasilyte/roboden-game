package sim

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/serverapi"
)

type httpServer struct {
	addr string
	busy int64

	resp          simulationResponse
	EventRequest  gsignal.Event[simulationRequest]
	EventShutdown gsignal.Event[error]
}

type simulationRequest struct {
	Config serverapi.LevelConfig `json:"config"`

	Actions []serverapi.PlayerAction `json:"actions"`
}

type simulationResponse struct {
	Err     string                `json:"err"`
	Results serverapi.GameResults `json:"results"`
}

func newHTTPServer(addr string) *httpServer {
	return &httpServer{addr: addr}
}

func (s *httpServer) Start() {
	http.HandleFunc("/roboden-sim/run", s.handleRun)
	http.HandleFunc("/roboden-sim/status", s.handleStatus)

	err := http.ListenAndServe(s.addr, nil)
	s.EventShutdown.Emit(err)
}

func (s *httpServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok"}`))
}

func (s *httpServer) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	var data simulationRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(data)
	if !s.prepareData(&data) {
		http.Error(w, "data validation error", http.StatusBadRequest)
		return
	}

	if !atomic.CompareAndSwapInt64(&s.busy, 0, 1) {
		http.Error(w, "already executing a simulation", http.StatusTooManyRequests)
		return
	}
	defer func() {
		atomic.StoreInt64(&s.busy, 0)
	}()

	s.resp = simulationResponse{}
	start := time.Now()
	s.EventRequest.Emit(data)
	elapsed := time.Since(start)
	log.Printf("the simulation completed in %v", elapsed)

	responseData, err := json.Marshal(s.resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseData)
}

func (s *httpServer) prepareData(req *simulationRequest) bool {
	req.Config.ExecMode = serverapi.ExecuteSimulation

	switch gamedata.Mode(req.Config.GameMode) {
	case gamedata.ModeClassic:
		req.Config.EnemyBoss = true
		return true

	default:
		// TODO.
		return false
	}
}
