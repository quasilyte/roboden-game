package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	var args arguments
	flag.StringVar(&args.output, "o", "",
		"the output binary name")
	flag.StringVar(&args.commit, "commit", "",
		"a commit hash")
	flag.BoolVar(&args.wasm, "wasm", false,
		"whether we're building for wasm target")
	flag.Parse()

	commit := args.commit
	if commit == "" {
		out, err := exec.Command("git", "rev-parse", "HEAD").CombinedOutput()
		if err != nil {
			panic(err)
		}
		commit = strings.TrimSpace(string(out))
	}

	ldFlags := []string{
		fmt.Sprintf("-X 'main.CommitHash=%s'", commit),
		"-X 'main.DefaultServerAddr=https://quasilyte.tech/roboden/api'",
		"-s -w",
	}
	goFlags := []string{
		"build",
		"--ldflags", strings.Join(ldFlags, " "),
		"--trimpath",
		"-v",
		"-o", args.output,
		"./cmd/game",
	}
	cmd := exec.Command("go", goFlags...)
	cmd.Env = append([]string{}, os.Environ()...) // Copy env slice
	if args.wasm {
		cmd.Env = append(cmd.Env, "GOOS=js")
		cmd.Env = append(cmd.Env, "GOARCH=wasm")
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("%v: %s", err, out))
	}
}

type arguments struct {
	wasm   bool
	commit string
	output string
}
