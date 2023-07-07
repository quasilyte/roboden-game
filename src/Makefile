COMMIT_HASH=`git rev-parse HEAD`

.PHONY: server runsim wasm itchio-wasm

server:
	go build -ldflags="-s -w -X 'main.CommitHash=$(COMMIT_HASH)'" -trimpath -o server ./cmd/server

runsim:
	go build -ldflags="-s -w -X 'main.CommitHash=$(COMMIT_HASH)'" -trimpath -o runsim ./cmd/runsim

wasm:
	GOARCH=wasm GOOS=js go build -ldflags="-s -w" -trimpath -o ../_web/main.wasm ./cmd/game

itchio-wasm: wasm
	cd ../_web && \
		rm -rf roboden_data && \
		cp -r ../roboden_data roboden_data && \
		mkdir -p ../bin && \
		rm -f ../bin/roboden.zip && \
		zip ../bin/roboden.zip -r main.wasm index.html wasm_exec.js roboden_data
