package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/lorenyeung/go-folder-create/models"

	logrus "github.com/sirupsen/logrus"
)

var flags models.Flags
var Log = logrus.New()
var GitCommit string
var Version string

type Flag struct {
}

func main() {
	flag := SetFlags(true)
	SetLogger(flag.LogLevelVar)
	preFolder := flag.DefaultFolderVar
	if flag.VersionVar {
		PrintVersion()
		os.Exit(0)
	}
	if flag.StructureVar == "" {
		Log.Panic("Please specify a tree structure file")
	}
	err := os.Mkdir(preFolder, 0755)
	CheckErr(err, false, "create parent folder failed", Trace())
	f, err := os.Open(flag.StructureVar)
	CheckErr(err, true, "read tree structure file failed", Trace())

	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		var path string
		if strings.Contains(line, "├── ") || strings.Contains(line, "└── ") {
			prefix := strings.Split(line, "── ") //removed ├── so now we have the full file/folder path
			Log.Debug("stripped prefix of ├── and └──:", prefix[1])
			path = prefix[1]

			if strings.Contains(path, " -> ") {
				Log.Trace("has symlink, stripping")
				symlink := strings.Split(prefix[1], " -> ")
				Log.Debug("stripped symlink:", symlink[0])
				path = symlink[0]

			}
		}
		//attempt at file detection via . existence of last part
		lastpart := strings.Split(path, "/")
		if strings.Contains(lastpart[len(lastpart)-1], ".") {
			Log.Info("file:", path)
			d1 := []byte("hello\ngo\n")
			err := os.WriteFile(preFolder+"/"+path, d1, 0644)
			CheckErr(err, false, "Failed to create file:"+path, Trace())
		} else {
			Log.Info("folder:", path)
			err := os.Mkdir(preFolder+"/"+path, 0755)
			CheckErr(err, false, "Failed to create folder:"+path, Trace())
		}
		Log.Trace("final path:", path, " orginal:", line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

//SetLogger sets logger settings
func SetLogger(logLevelVar string) {
	level, err := logrus.ParseLevel(logLevelVar)
	if err != nil {
		level = logrus.InfoLevel
	}
	Log.SetLevel(level)
	Log.SetReportCaller(true)
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.QuoteEmptyFields = true
	customFormatter.FullTimestamp = true
	customFormatter.DisableColors = false
	customFormatter.CallerPrettyfier = func(f *runtime.Frame) (string, string) {
		repopath := strings.Split(f.File, "/")
		function := strings.Replace(f.Function, "github.com/lorenyeung/go-folder-create", "", -1)
		return fmt.Sprintf("%s", function), fmt.Sprintf(" %s:%d", repopath[len(repopath)-1], f.Line)
	}
	//customFormatter.DisableSorting = true
	Log.SetFormatter(customFormatter)

	Log.Info("Log level set at ", level)
}

//Check logger for errors
func CheckErr(e error, panicCheck bool, logs string, trace models.TraceData) error {
	if e != nil && panicCheck {
		Log.Error(logs, " failed with error:", e, " ", trace.Fn, " on line:", trace.Line)
		panic(e)
	}
	if e != nil && !panicCheck {
		Log.Warn(logs, " failed with error:", e, " ", trace.Fn, " on line:", trace.Line)
		return e
	}
	return nil
}

func PrintVersion() {
	fmt.Println("Current build version:", GitCommit, "Current Version:", Version)
	Log.Info("Current build version:", GitCommit, "Current Version:", Version)
}

//SetFlags function
func SetFlags(init bool) models.Flags {
	if init {
		flag.StringVar(&flags.DefaultFolderVar, "folder", "folder-create", "Default parent folder to create your sub folders in")
		flag.StringVar(&flags.LogLevelVar, "log", "INFO", "Order of Severity: TRACE, DEBUG, INFO, WARN, ERROR, FATAL, PANIC")
		flag.StringVar(&flags.StructureVar, "structure", "", "tree file structure")
		flag.BoolVar(&flags.VersionVar, "v", false, "Print the current version and exit")
		flag.Parse()
	}
	return flags
}

//Trace get function data
func Trace() models.TraceData {
	var trace models.TraceData
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		Log.Warn("Failed to get function data")
		return trace
	}

	fn := runtime.FuncForPC(pc)
	trace.File = file
	trace.Line = line
	trace.Fn = fn.Name()
	return trace
}
