package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	args := parseCLIArgs()

	var l *fileLogger
	if args.logFile == "" {
		l = &fileLogger{f: os.Stderr}
	} else {
		f, err := os.Create(args.logFile)
		if err != nil {
			panic(err)
		}
		l = &fileLogger{
			f:        f,
			filename: args.logFile,
		}
	}

	mux := http.NewServeMux()
	config := serverConfig{
		runsimFolder: args.simulatorsFolder,
		httpHandler:  mux,
		dataFolder:   args.dataFolder,
		logger:       l,
		metricsFile:  args.metricsFile,
	}
	server := newAPIServer(config)

	if err := server.InitDatabases(); err != nil {
		panic(err)
	}
	if err := server.Preload(); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		server.BackgroundTask()
		wg.Done()
	}()

	h := newRequestHandler(server)
	httpServer := &http.Server{
		Addr:           args.listenAddr,
		Handler:        server,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1024 * 32,
	}

	mux.HandleFunc("/version", server.NewHandler(h.HandleVersion))
	mux.HandleFunc("/get-player-board", server.NewHandler(h.HandleGetPlayerBoard))
	mux.HandleFunc("/save-player-score", server.NewHandler(h.HandleSavePlayerScore))

	l.Info("starting server, listenning to %s", args.listenAddr)

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM)
	wg.Add(1)
	go func() {
		<-exitChan
		server.Stop()
		l.Info("caught the termination signal...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			l.Error("shutdown error: %v", err)
		}
		wg.Done()
	}()

	if err := httpServer.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
	wg.Wait()
	l.Info("the program is about to exit")
}

type cliArguments struct {
	listenAddr       string
	dataFolder       string
	metricsFile      string
	logFile          string
	simulatorsFolder string
}

func parseCLIArgs() *cliArguments {
	var args cliArguments

	flag.StringVar(&args.listenAddr, "listen", ":8080",
		"net listen address")
	flag.StringVar(&args.simulatorsFolder, "simulators-folder", "",
		"where to find roboden game simulators for replay validation")
	flag.StringVar(&args.dataFolder, "data-folder", "",
		"path to a sqlite databases folder")
	flag.StringVar(&args.metricsFile, "metrics", "metrics.json",
		"where to periodically dump server metrics")
	flag.StringVar(&args.logFile, "log", "",
		"write server logs to this file; stderr if empty")

	flag.Parse()

	return &args
}
