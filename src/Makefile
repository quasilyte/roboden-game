wasm:
	GOARCH=wasm GOOS=js go build -ldflags="-s -w" -o ../_web/main.wasm cmd/game/main.go

itchio-wasm:
	GOARCH=wasm GOOS=js go build -ldflags="-s -w" -o ../_web/main.wasm cmd/game/main.go
	cd ../_web && zip ../bin/roboden.zip main.wasm index.html wasm_exec.js
