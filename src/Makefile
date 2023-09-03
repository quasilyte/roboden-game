COMMIT_HASH=`git rev-parse HEAD`

.PHONY: server serverutil runsim wasm itchio-wasm steam-release

steam-release:
	rm -rf ../steam_dist
	mkdir -p ../steam_dist/linux_x86-64 ../steam_dist/windows_x86-32 ../steam_dist/windows_x86-64 ../steam_dist/out
	go run ./cmd/builder -o ../steam_dist/linux_x86-64/roboden -steam -goos=linux -goarch=amd64
	go run ./cmd/builder -o ../steam_dist/windows_x86-32/roboden.exe -steam -goos=windows -goarch=386
	go run ./cmd/builder -o ../steam_dist/windows_x86-64/roboden.exe -steam -goos=windows -goarch=amd64
	cp -r ../roboden_data ../steam_dist/linux_x86-64/roboden_data
	cp -r ../roboden_data ../steam_dist/windows_x86-32/roboden_data
	cp -r ../roboden_data ../steam_dist/windows_x86-64/roboden_data
	cp ../_steam/steam_api.dll ../steam_dist/windows_x86-32/
	cp ../_steam/steam_api64.dll ../steam_dist/windows_x86-64/
	cd ../steam_dist/linux_x86-64 && zip -r ../out/roboden_linux.zip *
	cd ../steam_dist/windows_x86-32 && zip -r ../out/roboden_windows32.zip *
	cd ../steam_dist/windows_x86-64 && zip -r ../out/roboden_windows64.zip *
	@echo "[OK] all done"

server:
	go build -ldflags="-s -w -X 'main.CommitHash=$(COMMIT_HASH)'" -trimpath -o server ./cmd/server

serverutil:
	go build -ldflags="-s -w -X 'main.CommitHash=$(COMMIT_HASH)'" -trimpath -o serverutil ./cmd/serverutil

runsim:
	go build -ldflags="-s -w -X 'main.CommitHash=$(COMMIT_HASH)'" -trimpath -o runsim_x ./cmd/runsim

wasm:
	GOARCH=wasm GOOS=js go build -ldflags="-s -w" -trimpath -o ../_web/main.wasm ./cmd/game

itchio-wasm: wasm
	cd ../_web && \
		rm -rf roboden_data && \
		cp -r ../roboden_data roboden_data && \
		mkdir -p ../bin && \
		rm -f ../bin/roboden.zip && \
		zip ../bin/roboden.zip -r main.wasm index.html wasm_exec.js roboden_data
