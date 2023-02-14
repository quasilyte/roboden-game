![logo](_metadata/logo.png)

The **online demo** and **release binaries** can be found here: <https://quasilyte.itch.io/roboden>.

## Game Overview

Two robotic life forms collided, and only one will remain. Can you lead the drone-producing colony to victory or will you face defeat?

This game allows you to build units and bases, harvest resources, and explore the hostile world without the direct control you're used to having in most RTS games. Instead of giving a direct unit order, you manipulate the colony's priorities and let it decide what needs to be done (and how it should be done).

This game is a "My First Game Jam: Winter 2023" submission. It was created during a single week of development by a team of two.

Features:

* Asymmetrical RTS gameplay
* Unique base and units control system
* Neat pixel art graphics
* Randomized stage generation with customizations
* Unit combining system for higher tier units
* Easy to learn, hard to master game process
* Interactive in-game tutorial
* 14 different drones divided into 3 tiers
* 4 drone factions, each with their own bonuses

If you're playing a browser version of the game, please use Chrome or some other browser that has good wasm support (you may have performance issues in Firefox). If possible, prefer a native build instead; you'll get a smooth 60fps experience this way.

![screenshot](_metadata/screenshot.png)

## How to Run

```bash
git clone https://github.com/quasilyte/roboden-game.git
cd roboden-game/src
go run ./cmd/game
```

> You will need a [go](https://go.dev/) 1.18+ toolchain in order to build this game.

You may need to install [Ebitengine dependencies](https://ebitengine.org/en/documents/install.html#Installing_dependencies):

```bash
# For Debian/Ubuntu
$ sudo apt install libc6-dev libglu1-mesa-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config
```

If you want to build a game for a different platform, use Go cross-compilation:

```bash
GOOS=windows go build -o ../bin/decipherism.exe ./cmd/game
```

This game is tested on these targets:

* windows/amd64
* linux/amd64
* darwin/arm64
* js/wasm
