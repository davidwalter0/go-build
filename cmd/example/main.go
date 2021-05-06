package main

import (
	"fmt"
	"os"

	"github.com/davidwalter0/go-build"
)

func init() {
	fmt.Fprintf(os.Stderr, "\nBuildinfo for %s\n\n", os.Args[0])
	build.BuildInfo()
}
func main() {
}
