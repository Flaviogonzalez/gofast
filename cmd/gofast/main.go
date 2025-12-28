package main

import (
	"os"

	"github.com/flaviogonzalez/gofast/internal/cmd"
)

func main() {
	cmd.Run(os.Stderr, os.Stdout, os.Args[1:], os.Stdin)
}
