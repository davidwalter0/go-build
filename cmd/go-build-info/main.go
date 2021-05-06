// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2021-04-23 12:34:34

// === Go Playground ===
// Execute the snippet with Ctl-Return
// Provide custom arguments to compile with Alt-Return
// Remove the snippet completely with its dir and all files M-x `go-playground-rm`

// https://github.com/src-d/go-git/pull/1096/files
// check merge base

// build ignore
// go:run go build -o generate-buildinfo

// snippet of code @ 2020-08-11 13:57:45

// https://stackoverflow.com/questions/38517593/relative-imports-in-go
/*
   mkdir pkg
   cd pkg
   go mod init pkg
*/
// === Go Playground ===
// Execute the snippet with Ctl-Return
// Provide custom arguments to compile with Alt-Return
// Remove the snippet completely with its dir and all files M-x `go-playground-rm`

// https://gist.github.com/davidwalter0/60f41b53732656c5c546cc8b0a739d11

// x/vgo: thoughts on selection criteria for non-versioned go repo

// Run git branch version selection for non-vgo versioned repository
// as if in a wayback machine.

// Use the dependency's commit ctime as the wayback ctime.
// Commits prior to the wayback ctime in sub-dependencies are eligible.
// Use the newest commit prior to the wayback ctime.

// Rough sketch of wayback idea to find a candidate ctime of a git
// commit

// This would need to be integrated into vgo logic and of course
// non-git vc methods.

// Add tooling from go-git examples to remove the external dependency

// Select some arbitrary ctime for the wayback time

// export ctime='2017-09-04 19:43:36 +0300';
// vgo run main.go /go/src/github.com/go-git/go-git/v5 "${ctime}"

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// VersionArgs for cli
type VersionArgs struct {
	WriteRaceDetector bool `json:"race-detector" default:"true" doc:"enable race.go"`
}

var (
	versionArgs = &VersionArgs{}
)

// Open an existing repository in a specific folder.
func main() {
	fmt.Println()
	var When time.Time = time.Now().Add(24 * time.Hour)
	var err error
	var now = time.Now().Format(Layout)

	var topLevel string
	var ctime = "now"

	topLevel, err = Toplevel(".")

	if ctime == "now" {
		ctime = now
	}

	if When, err = time.Parse(Layout, ctime); err != nil {
		panic(err)
	}

	// We instance a new repository targeting the given topLevel (the .git folder)
	var repo *git.Repository
	repo, err = git.PlainOpen(topLevel)
	if err != nil {
		panic(err)
	}
	var ref *plumbing.Reference
	// ... retrieving the HEAD reference
	ref, err = repo.Head()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	workTree, err := repo.Worktree()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	status, err := workTree.Status()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// https://github.com/src-d/go-git/issues/923

	dirty := len(status) != 0

	var cIter object.CommitIter
	var requireTagInfo bool = true
	cIter, err = repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Find a tagged commit
	var wf = NewWayback(repo, requireTagInfo, When, cIter)
	var tagged *object.Commit
	var untagged *object.Commit
	var c *object.Commit
	var tag string

	c, tag, err = wf.Find(nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)

	}
	fmt.Printf("%-8.8s %-32.32s %-12.12s\n", "Hash", "Commit Time", "Tag")
	if c != nil {
		fmt.Printf("Tagged   %-32.32s %-12.12s %-20.20s %s-%s-%12.12s\n", c.Hash, tag, c.Committer.When,
			strings.TrimSpace(tag), vgoFormat(fmt.Sprintf("%-14.14s", c.Committer.When)), c.Hash)
		tagged = c
	}
	cIter, err = repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Find any newer commit
	wf = NewWayback(repo, !requireTagInfo, When, cIter)
	wf.Debug = false
	c, _, err = wf.Find(tagged)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if c != nil {
		var vgoUntag string
		if len(vgoUntag) == 0 {
			vgoUntag = "v0.0.0"
		}
		fmt.Printf("Untagged %-32.32s %-12.12s %-20.20s %s-%s-%12.12s\n", c.Hash, "", c.Committer.When,
			strings.TrimSpace(vgoUntag), vgoFormat(fmt.Sprintf("%-14.14s", c.Committer.When)), c.Hash)
		untagged = c
	}
	switch {
	case untagged == nil && tagged == nil:
		fmt.Println("unable to find commit")
		os.Exit(1)
	case untagged != nil && tagged == nil:
		WriteBuildInfo("", untagged, dirty, false)
	case untagged == nil && tagged != nil:
		WriteBuildInfo(tag, tagged, dirty, true)
	case untagged != nil && tagged != nil && untagged.Committer.When.After(tagged.Committer.When):
		WriteBuildInfo("", untagged, dirty, false)
	default:
		// case untagged != nil && tagged != nil && tagged. newer than untagged
		WriteBuildInfo(tag, tagged, dirty, true)
	}
}

// FileExists test for file
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// WriteBuildInfo create a new build.go
func WriteBuildInfo(tag string, c *object.Commit, dirty, isTagged bool) {
	fmt.Printf("%-8.8s %-32.32s %-12.12s %-20.20s\n",
		func() string {
			if isTagged {
				return "Tagged"
			}
			return "Untagged"
		}(),
		c.Hash, tag, c.Committer.When)

	var hash, when, now = c.Hash.String(), c.Committer.When.Format(Layout), time.Now().Format(Layout)
	var short = hash[:7]

	fmt.Println(hash, tag, dirty, when, now)

	var goos = os.Getenv("GOOS")
	var arch = os.Getenv("GOARCH")
	var ldflags map[string]string = make(map[string]string)

	os.Setenv("BUILD_INFO_GIT_COMMIT_DATE", when)
	os.Setenv("BUILD_INFO_GIT_REVISION", func() string {
		var revision = tag
		var suffix string
		if dirty {
			suffix = "-dirty"
		}
		if len(revision) == 0 {
			revision = short
		}
		return revision + suffix
	}())
	os.Setenv("BUILD_INFO_GIT_VERSION", hash)
	os.Setenv("BUILD_INFO_GO_ARCH_BUILT_ON", runtime.GOARCH)
	os.Setenv("BUILD_INFO_GO_ARCH_BUILT_FOR", func() string {
		if len(arch) > 0 {
			return arch
		}
		return runtime.GOARCH
	}())
	os.Setenv("BUILD_INFO_GO_BUILT_DATE", now)
	os.Setenv("BUILD_INFO_GO_COMPILER", runtime.Compiler)
	os.Setenv("BUILD_INFO_GO_OS_BUILT_ON", runtime.GOOS)
	os.Setenv("BUILD_INFO_GO_OS_BUILT_FOR", func() string {
		if len(goos) > 0 {
			return goos
		}
		return runtime.GOOS
	}())

	var raceDetector string = "false"
	for _, arg := range os.Args {
		if arg == "--race" || arg == "-race" {
			raceDetector = "true"
			break
		}
	}
	os.Setenv("BUILD_INFO_GO_RACE_DETECTOR", raceDetector)
	os.Setenv("BUILD_INFO_GO_VERSION", runtime.Version())
	for _, e := range os.Environ() {
		if len(e) > len("BUILD_INFO") && e[:len("BUILD_INFO")] == "BUILD_INFO" {
			split := strings.Split(e, "=")
			if len(split) > 1 {
				lhs, rhs := split[0], split[1]
				fmt.Printf("%-32s %s\n", lhs, rhs)
			}
		}
	}
	var BUILD_INFO_GIT_COMMIT_DATE = os.Getenv("BUILD_INFO_GIT_COMMIT_DATE")
	var BUILD_INFO_GIT_REVISION = os.Getenv("BUILD_INFO_GIT_REVISION")
	var BUILD_INFO_GIT_VERSION = os.Getenv("BUILD_INFO_GIT_VERSION")
	var BUILD_INFO_GO_ARCH_BUILT_ON = os.Getenv("BUILD_INFO_GO_ARCH_BUILT_ON")
	var BUILD_INFO_GO_ARCH_BUILT_FOR = os.Getenv("BUILD_INFO_GO_ARCH_BUILT_FOR")
	var BUILD_INFO_GO_BUILT_DATE = os.Getenv("BUILD_INFO_GO_BUILT_DATE")
	var BUILD_INFO_GO_COMPILER = os.Getenv("BUILD_INFO_GO_COMPILER")
	var BUILD_INFO_GO_OS_BUILT_ON = os.Getenv("BUILD_INFO_GO_OS_BUILT_ON")
	var BUILD_INFO_GO_OS_BUILT_FOR = os.Getenv("BUILD_INFO_GO_OS_BUILT_FOR")
	var BUILD_INFO_GO_RACE_DETECTOR = os.Getenv("BUILD_INFO_GO_RACE_DETECTOR")
	var BUILD_INFO_GO_VERSION = os.Getenv("BUILD_INFO_GO_VERSION")

	ldflags["BUILD_INFO_GIT_COMMIT_DATE"] = BUILD_INFO_GIT_COMMIT_DATE
	ldflags["BUILD_INFO_GIT_REVISION"] = BUILD_INFO_GIT_REVISION
	ldflags["BUILD_INFO_GIT_VERSION"] = BUILD_INFO_GIT_VERSION
	ldflags["BUILD_INFO_GO_ARCH_BUILT_ON"] = BUILD_INFO_GO_ARCH_BUILT_ON
	ldflags["BUILD_INFO_GO_ARCH_BUILT_FOR"] = BUILD_INFO_GO_ARCH_BUILT_FOR
	ldflags["BUILD_INFO_GO_BUILT_DATE"] = BUILD_INFO_GO_BUILT_DATE
	ldflags["BUILD_INFO_GO_COMPILER"] = BUILD_INFO_GO_COMPILER
	ldflags["BUILD_INFO_GO_OS_BUILT_ON"] = BUILD_INFO_GO_OS_BUILT_ON
	ldflags["BUILD_INFO_GO_OS_BUILT_FOR"] = BUILD_INFO_GO_OS_BUILT_FOR
	ldflags["BUILD_INFO_GO_RACE_DETECTOR"] = BUILD_INFO_GO_RACE_DETECTOR
	ldflags["BUILD_INFO_GO_VERSION"] = BUILD_INFO_GO_VERSION
	var Package = func() string {
		var e = os.Getenv("GO_PACKAGE")
		if len(e) > 0 {
			return e
		}
		return buildInfoPackagePath
	}()
	var ldargsFlag string = "-ldflags"
	var ldargs string
	for k, v := range ldflags {
		ldargs += fmt.Sprintf(` -X %s.%s=%s`, Package, k, v)
	}
	ldargs = "" + strings.TrimSpace(ldargs) + ""

	var args []string = []string{os.Args[1], os.Args[2]}

	if len(os.Args) > 2 {

		var in bool
		for _, arg := range os.Args[3:] {
			arg = strings.TrimSpace(arg)
			switch in {
			case true:
				ldargs += fmt.Sprintf(" %s", arg)
			case false:
				if in = isLDFlag(arg); in {
					continue // skip this flag inplace, and append it to to our args
				}
				args = append(args, fmt.Sprintf("%s", arg))
			}
		}
		args = append(args, ldargsFlag)
		args = append(args, ldargs)
		log.Printf("\nBuilding with package arg %s and command line\n%s\n", Package, args)
		var err error
		var cmd *exec.Cmd = exec.Command(args[0], args[1:]...)
		var output bytes.Buffer
		cmd.Stdout = &output
		var errors bytes.Buffer
		cmd.Stderr = &errors
		err = cmd.Run()
		log.Println(errors.String())
		log.Println(output.String())
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}
}

func isLDFlag(arg string) (in bool) {
	var ldflags string
	if len(arg) > 5 {
		if arg[0] == '-' && arg[1] == '-' {
			ldflags = arg[2:]
		} else {
			if arg[0] == '-' {
				ldflags = arg[1:]
			}
		}
		if ldflags == "ldflags" {
			in = true
		}
	}
	return
}

func vgoFormat(text string) string {
	text = strings.Replace(text, ":", "", -1)
	text = strings.Replace(text, "-", "", -1)
	text = strings.Replace(text, " ", "", -1)
	return text
}

const (
	// Layout format spec for parsing
	Layout = "2006.01.02.15.04.05.-0700"
)

var (
	// ErrNotFound error
	ErrNotFound = fmt.Errorf("Reference not found")
)

// Wayback repo entry info
type Wayback struct {
	Debug bool
	*git.Repository
	RequireTagInfo bool
	When           time.Time
	CommitIter     object.CommitIter
}

// Taginfo for reporting build info generation detail
type Taginfo struct {
	Tag  string
	Hash plumbing.Hash
	When time.Time
}

// ByCommitTimeTagInfo sortable slice of Taginfo
type ByCommitTimeTagInfo []Taginfo

func (tags ByCommitTimeTagInfo) Len() int           { return len(tags) }
func (tags ByCommitTimeTagInfo) Less(i, j int) bool { return tags[i].When.Before(tags[j].When) }
func (tags ByCommitTimeTagInfo) Swap(i, j int)      { tags[i], tags[j] = tags[j], tags[i] }

// NewWayback create a Wayback object
func NewWayback(r *git.Repository, requireTagInfo bool, when time.Time, commitIter object.CommitIter) *Wayback {
	return &Wayback{Repository: r, RequireTagInfo: requireTagInfo, When: when, CommitIter: commitIter}
}

// Find and return commit info and tag if present
func (wayback *Wayback) Find(tagged *object.Commit) (c *object.Commit, tag string, err error) {
	defer wayback.CommitIter.Close()
	if wayback.RequireTagInfo {
		return wayback.FindFirstTag()
	}
	return wayback.FindFirst(tagged)
}

// FindFirstNoTag commit prior to the wayback time without a tag
func (wayback *Wayback) FindFirstNoTag() (c *object.Commit, tag string, err error) {
	for {
		if c, err = wayback.CommitIter.Next(); err == io.EOF {
			break
		}
		if err != nil {
			c = nil
			return
		}
		if c.Committer.When.Before(wayback.When) {
			return
		}
	}
	return
}

// FindFirst commit prior to the wayback time
func (wayback *Wayback) FindFirst(tagged *object.Commit) (c *object.Commit, tag string, err error) {
	for {
		if c, err = wayback.CommitIter.Next(); err == io.EOF {
			break
		}
		if err != nil {
			c = nil
			return
		}
		if c.Committer.When.Before(wayback.When) {
			return
		}
	}
	return
}

// FindFirstTag commit prior to the wayback time
func (wayback *Wayback) FindFirstTag() (c *object.Commit, tag string, err error) {
	var byCommitTimeTagInfo ByCommitTimeTagInfo
	var tags storer.ReferenceIter
	var ref *plumbing.Reference

	if tags, err = wayback.Repository.Tags(); err != nil {
		return
	}

	byCommitTimeTagInfo = make(ByCommitTimeTagInfo, 0)
	for {
		if ref, err = tags.Next(); err == io.EOF {
			err = nil
			tag = "v0.0.0"
			break
		}
		if err != nil {
			c = nil
			return
		}
		if c, err = wayback.Repository.CommitObject(ref.Hash()); err != nil {
			c = nil
			return
		}
		byCommitTimeTagInfo = append(byCommitTimeTagInfo,
			Taginfo{Tag: ref.Name().Short(), Hash: ref.Hash(), When: c.Committer.When})
	}
	sort.Sort(sort.Reverse(byCommitTimeTagInfo))
	var tagInfo Taginfo
	if wayback.Debug {
		fmt.Printf("Searching for tagged commit newer than %-32.32s\n",
			wayback.When)
		fmt.Printf("%-8.8s %-32.32s %-12.12s\n", "Hash", "Commit Time", "Tag")
	}
	for _, tagInfo = range byCommitTimeTagInfo {
		if c, err = wayback.Repository.CommitObject(tagInfo.Hash); err != nil {
			c = nil
			return
		}
		if wayback.Debug {
			fmt.Printf("%-8.8s %-32.32s %-12.12s\n", tagInfo.Hash, c.Committer.When, tagInfo.Tag)
		}
		if c.Committer.When.Before(wayback.When) {
			tag = tagInfo.Tag
			err = nil
			return
		}
	}
	//	err = ErrNotFound
	return
}

// Tag returns status of flag found and tag value
func Tag(repo *git.Repository) (tag string, isTag bool, err error) {
	var ref *plumbing.Reference
	var tags storer.ReferenceIter

	if ref, err = repo.Head(); err != nil {
		return
	}

	if tags, err = repo.Tags(); err != nil {
		return
	}

	if err = tags.ForEach(func(_ref *plumbing.Reference) error {
		if _ref.Hash().String() == ref.Hash().String() {
			isTag = true
			tag = _ref.Name().Short()
			return nil
		}
		return nil
	}); err != nil {
		return
	}
	// err = ErrNotFound
	return
}

// common tooling from examples

// CheckArgs should be used to ensure the right command line arguments are
// passed before executing an example.
func CheckArgs(help string, arg ...string) {
	if len(os.Args) < len(arg)+1 {
		UseMessage(help, "Usage:", "%s %s", os.Args[0], strings.Join(arg, " "))
		os.Exit(1)
	}
}

// CheckIfError naively panics if an error is not nil.
func CheckIfError(err error, help string, arg ...string) {
	if err == nil {
		return
	}
	var prefix = fmt.Sprintf(`
Failed command was

       %s

`, strings.Join(os.Args, " "))

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	UseMessage(help, prefix, `
Usage:

       %s %s

`, os.Args[0], strings.Join(arg, " "))
	os.Exit(1)
}

// Info should be used to describe the example commands that are about to run.
func Info(format string, versionArgs ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, versionArgs...))
}

// Warning should be used to display a warning
func Warning(format string, versionArgs ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, versionArgs...))
}

// UseMessage should be used to display a Error
func UseMessage(help, prefix, format string, versionArgs ...interface{}) {
	fmt.Printf("\n%s \x1b[36;1m%s\x1b[0m\n%s\n", prefix, fmt.Sprintf(format, versionArgs...), help)
}

const metadataDir = "/.git"

// Toplevel of a git repository path without the metadataDir suffix.
func Toplevel(path string) (string, error) {
	toplevel, err := DetectGitPath(".")
	if err != nil {
		return "", err
	}
	if len(toplevel) > len(metadataDir) {
		toplevel = toplevel[:len(toplevel)-(len(metadataDir))]
	}
	return toplevel, err
}

// DetectGitPath finds if path is in a git repo tree
func DetectGitPath(path string) (string, error) {
	// normalize the path
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	for {
		fi, err := os.Stat(filepath.Join(path, ".git"))
		if err == nil {
			if !fi.IsDir() {
				return "", fmt.Errorf(".git exist but is not a directory")
			}
			return filepath.Join(path, ".git"), nil
		}
		if !os.IsNotExist(err) {
			// unknown error
			return "", err
		}

		// detect bare repo
		ok, err := IsGitDir(path)
		if err != nil {
			return "", err
		}
		if ok {
			return path, nil
		}
		parent := filepath.Dir(path)
		if parent == path {
			return "", fmt.Errorf(".git not found")
		}
		path = parent
	}
}

// IsGitDir tests for git repo
func IsGitDir(path string) (bool, error) {
	markers := []string{"HEAD", "objects", "refs"}

	for _, marker := range markers {
		_, err := os.Stat(filepath.Join(path, marker))
		if err == nil {
			continue
		}
		if !os.IsNotExist(err) {
			// unknown error
			return false, err
		}
		return false, nil
	}

	return true, nil
}
