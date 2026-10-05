package main

import (
	"os"

	"github.com/erikkubica/gmcp/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
