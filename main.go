package main

import (
	"github.com/xbt573/flood-social-rep/cmd"
	"os"
)

func main() {
	err := cmd.Run()
	if err != nil {
		os.Exit(1)
	}
}
