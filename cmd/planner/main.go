package main

import (
	"fmt"
	"os"

	"github.com/neatflowcv/worker/internal/app/plannercli"
)

func main() {
	err := plannercli.Run()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}
