package main

import (
	"os"

	"stitchforce-server/cmd/license/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}