COMMIT_HASH=`git rev-parse HEAD`

.PHONY: server runsim wasm itchio-wasm

server:
	go build -ldflags="-s -w -X 'main.CommitHash=$(COMMIT_HASH)'" -trimpath -o server ./cmd/server

runsim:
	go build -ldflags="-s -w -X 'main.CommitHash=$(COMMIT_HASH)'" -trimpath -o runsim ./cmd/runsim

wasm:
	GOARCH=wasm GOOS=js go build -ldflags="-s -w" -trimpath -o ../_web/main.wasm cmd/game/main.go

itchio-wasm: wasm
	cd ../_web && zip ../bin/roboden.zip main.wasm index.html wasm_exec.js
