package main

import (
	"fmt"
	"os"

	"github.com/neatflowcv/worker/internal/app/cli"
)

func main() {
	err := cli.Run()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}
