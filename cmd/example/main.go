package main

import (
	"fmt"
	"os"

	"github.comcast.com/dwalte022/go-build"
)

func init() {
	fmt.Fprintf(os.Stderr, "\nBuildinfo for %s\n\n", os.Args[0])
	build.BuildInfo()
}
func main() {
}
