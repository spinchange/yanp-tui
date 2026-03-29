package main

import (
	"fmt"
	"os"

	"github.com/spinchange/yanp-tui/internal/app"
	"github.com/spinchange/yanp-tui/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	if err := app.Run(os.Args[1:], cfg); err != nil {
		fmt.Fprintf(os.Stderr, "yanp: %v\n", err)
		os.Exit(1)
	}
}
