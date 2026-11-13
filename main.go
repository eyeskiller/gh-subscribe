package main

import (
	"os"

	"github.com/lubosberan/gh-subscribe/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
