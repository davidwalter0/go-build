package build

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
)

var Package = self()

func self() string {
	pc, _, _, _ := runtime.Caller(1)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	pkage := ""
	funcName := parts[pl-1]
	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
		pkage = strings.Join(parts[0:pl-2], ".")
	} else {
		pkage = strings.Join(parts[0:pl-1], ".")
	}
	return pkage
}

func init() {
	BuildInfo()
}

var BUILD_INFO_GIT_COMMIT_DATE string
var BUILD_INFO_GIT_REVISION string
var BUILD_INFO_GIT_VERSION string
var BUILD_INFO_GO_ARCH_BUILT_ON string
var BUILD_INFO_GO_ARCH_BUILT_FOR string
var BUILD_INFO_GO_BUILT_DATE string
var BUILD_INFO_GO_COMPILER string
var BUILD_INFO_GO_OS_BUILT_ON string
var BUILD_INFO_GO_OS_BUILT_FOR string
var BUILD_INFO_GO_RACE_DETECTOR string
var BUILD_INFO_GO_VERSION string

// XJSON writes a json encoded format of XInfo to the given io.Writer.
func XJSON(w io.Writer, info XInfo) error {
	enc := json.NewEncoder(w)
	return enc.Encode(info)
}

type XInfo struct {
	GitVersion    string `json:"version"`
	GitRevision   string `json:"revision"`
	GitCommitDate string `json:"commit-date"`
	Date          string `json:"build-date"`
	RaceDetector  string `json:"race-detector"`
	Arch          string `json:"arch"`
	OS            string `json:"os"`
	BuildArch     string `json:"build-arch"`
	BuildOS       string `json:"build-os"`
	Compiler      string `json:"compiler"`
	GoVersion     string `json:"go-version"`
}

func BuildInfo() {
	var info = XInfo{
		GitVersion:    BUILD_INFO_GIT_REVISION,
		GitRevision:   BUILD_INFO_GIT_VERSION,
		GitCommitDate: BUILD_INFO_GIT_COMMIT_DATE,
		Date:          BUILD_INFO_GO_BUILT_DATE,
		RaceDetector:  BUILD_INFO_GO_RACE_DETECTOR,
		BuildArch:     BUILD_INFO_GO_ARCH_BUILT_FOR,
		BuildOS:       BUILD_INFO_GO_OS_BUILT_ON,
		Arch:          runtime.GOARCH,
		OS:            runtime.GOOS,
		Compiler:      BUILD_INFO_GO_COMPILER,
		GoVersion:     BUILD_INFO_GO_VERSION,
	}

	fmt.Fprintf(os.Stderr, "%s : %s\n", Teal("Git Commit Date "), info.GitCommitDate)
	fmt.Fprintf(os.Stderr, "%s : %s\n", Teal("Git Revision    "), info.GitRevision)
	fmt.Fprintf(os.Stderr, "%s : %s\n", Teal("Git Version     "), info.GitVersion)
	fmt.Fprintf(os.Stderr, "%s : %s\n", Teal("Go Build ARCH   "), info.BuildArch)
	fmt.Fprintf(os.Stderr, "%s : %s\n", Teal("Go Build OS     "), info.BuildOS)
	fmt.Fprintf(os.Stderr, "%s : %s\n", Teal("Go Build Date   "), info.Date)
	fmt.Fprintf(os.Stderr, "%s : %s\n", Teal("Go ARCH         "), info.Arch)
	fmt.Fprintf(os.Stderr, "%s : %s\n", Teal("Go Build Date   "), info.Date)
	fmt.Fprintf(os.Stderr, "%s : %s\n", Teal("Go Compiler     "), info.Compiler)
	fmt.Fprintf(os.Stderr, "%s : %s\n", Teal("Go OS           "), info.OS)
	fmt.Fprintf(os.Stderr, "%s : %s\n", Teal("Go Race Detector"), info.RaceDetector)
	fmt.Fprintf(os.Stderr, "%s : %s\n", Teal("Go Version      "), info.GoVersion)
	//	XJSON(os.Stderr, info)
}

var (
	// Fail aliases Red text color function
	Fail = Red
	// 	Info  = Teal
	// 	Warn  = Yellow
	// 	Fatal = Red
	// 	Pass  = Green

)

var (
	Red     = Color("\033[1;31m%s\033[0m")
	Black   = Color("\033[1;30m%s\033[0m")
	Green   = Color("\033[1;32m%s\033[0m")
	Yellow  = Color("\033[1;33m%s\033[0m")
	Purple  = Color("\033[1;34m%s\033[0m")
	Magenta = Color("\033[1;35m%s\033[0m")
	Teal    = Color("\033[1;36m%s\033[0m")
	White   = Color("\033[1;37m%s\033[0m")
)

// Color returns a function that configures a scoped string colorizor
func Color(colorString string) func(...interface{}) string {
	if !color.NoColor {
		return func(versionArgs ...interface{}) string {
			return fmt.Sprintf(colorString, fmt.Sprint(versionArgs...))
		}
	}
	return func(versionArgs ...interface{}) string {
		return fmt.Sprint(versionArgs...)
	}
}
