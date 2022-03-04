package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
)

func main() {
	preFolder := "folder-create"
	if len(os.Args) <= 1 {
		fmt.Println("Please specify file as first argument")
		os.Exit(1)
	}
	err := os.Mkdir(preFolder, 0755)
	if err != nil {
		fmt.Println(err)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		//fmt.Println(scanner.Text())
		line := scanner.Text()
		var path string
		if strings.Contains(line, "├── ") || strings.Contains(line, "└── ") {
			prefix := strings.Split(line, "── ")
			//fmt.Println(prefix[1])
			//removed ├── so now we have the full file/folder path
			path = prefix[1]

			if strings.Contains(path, " -> ") {
				//fmt.Println(prefix[1], "has symlink, stripping")
				symlink := strings.Split(prefix[1], " -> ")
				path = symlink[0]

			}
		}
		//attempt at file detection
		lastpart := strings.Split(path, "/")
		if strings.Contains(lastpart[len(lastpart)-1], ".") {
			fmt.Println("file:", path)
			d1 := []byte("hello\ngo\n")
			err := os.WriteFile(preFolder+"/"+path, d1, 0644)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("folder:", path)
			err := os.Mkdir(preFolder+"/"+path, 0755)
			if err != nil {
				fmt.Println(err)
			}
		}
		//fmt.Println(path, line)
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
		function := strings.Replace(f.Function, "github.com/lorenyeung/go-orchestrate/", "", -1)
		return fmt.Sprintf("%s", function), fmt.Sprintf(" %s:%d", repopath[len(repopath)-1], f.Line)
	}
	customFormatter.DisableSorting = true
	customFormatter.SortingFunc = func(keys []string) {
		//doesnt do anything...
		keys = []string{"time", "level", "file", "func", "msg"}
		return
	}
	Log.SetFormatter(customFormatter)

	// Log.SetFormatter(&easy.Formatter{
	// 	TimestampFormat: "15:04:05 2006-01-02",
	// 	LogFormat:       "[%lvl%]: %time% - %fn% %msg%\n",
	// })

	if _, err := os.Stat(OrchestrateFlags.FolderVar + "log/"); os.IsNotExist(err) {
		fmt.Println("No log folder, generating under:" + OrchestrateFlags.FolderVar + "log/")
		err = os.MkdirAll(OrchestrateFlags.FolderVar+"log/", 0700)
		if err != nil {
			fmt.Println("Error generating log directory. Logs will not be shown:", err)
		}
	}

	file, err := os.OpenFile(OrchestrateFlags.FolderVar+"log/orchestrate-service.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		if OrchestrateFlags.ModeVar == "development" {
			mw := io.MultiWriter(os.Stdout, file)
			Log.SetOutput(mw)

		} else {
			Log.Out = file
		}
	} else {
		Log.Info("Failed to log to file, using default stderr:", err)
	}

	Log.Info("Log level set at ", level)
}
