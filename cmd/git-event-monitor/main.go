package main

import (
	"os"

	"github.com/luoliwoshang/git-event-monitor/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}