package main

import (
	"github.com/kapok/kapok/cmd/kapok/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
