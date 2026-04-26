package main

import (
	"os"

	"github.com/braincrew/agentsail/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
