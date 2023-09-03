package main

import (
	"log"

	"github.com/cespare/subcmd"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.SetFlags(0)

	cmds := []subcmd.Command{
		{
			Name:        "archive.extract",
			Description: "extract archived replay",
			Do:          makeMainFunc(cmdArchiveExtract),
		},

		{
			Name:        "version",
			Description: "print tool version info",
			Do:          makeMainFunc(cmdVersion),
		},
	}

	subcmd.Run(cmds)
}

func makeMainFunc(f func(args []string) error) func(args []string) {
	return func(args []string) {
		if err := f(args); err != nil {
			log.Fatalf("error: %v", err)
		}
	}
}
