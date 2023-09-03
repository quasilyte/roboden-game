package main

import (
	"fmt"
)

func cmdVersion(args []string) error {
	if CommitHash == "" {
		fmt.Printf("no version info\n")
	} else {
		fmt.Printf("serverutils version %s", CommitHash)
	}
	return nil
}
