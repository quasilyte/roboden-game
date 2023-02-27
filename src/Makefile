wasm:
	GOARCH=wasm GOOS=js go build -o ../_web/main.wasm cmd/game/main.go
	cd ../_web && zip ../bin/roboden.zip *
