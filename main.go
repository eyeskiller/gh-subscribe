package main

import (
	"os"

	"github.com/eyeskiller/gh-subscribe/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
