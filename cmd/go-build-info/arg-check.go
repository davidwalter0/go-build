package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/davidwalter0/go-build"
)

var buildInfoPackagePath = build.Package

func init() {
	log.Println("buildInfoPackagePath", buildInfoPackagePath)
	// os.Exit(0)

	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, Usage())
		os.Exit(1)
	}
	if !(filepath.Base(os.Args[1]) == "go" && (os.Args[2] == "build" || os.Args[2] == "install")) {
		fmt.Fprintln(os.Stderr, Usage())
		os.Exit(1)
	}
}

// Usage run arg info
func Usage() string {
	var args []string = func() []string {
		if len(os.Args) == 1 {
			return []string{}
		}
		return os.Args[1:]
	}()
	return fmt.Sprintf(`
Usage: %s go build [build args...]

Alt-Usage: %s go install [build args...]

e.g.

build:  %s /usr/bin/go build -ldflags '-w -s -X main.Apple="Pear"'

or

install:  %s /usr/bin/go install -ldflags '-w -s -X main.Apple="Pear"'

Requires a go build command to start

Command line provided was
%s
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], args)
}
