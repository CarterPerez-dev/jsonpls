// ===================
// © AngelaMos | 2026
// main.go
// ===================

package main

import (
	"os"

	"github.com/CarterPerez-dev/jsonpls/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
