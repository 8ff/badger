package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/8ff/tuna"
)

var Version string

var initFile string
var logFile string

func tlog(logType string, logData string) {
	if logType == "error" {
		fmt.Fprintf(os.Stderr, "%s [%s]% -s %s\n", time.Now().Format("02/01/2006_15:04:05.000000"), logType, "", logData)
	} else {
		fmt.Printf("%s [%s]% -s %s\n", time.Now().Format("02/01/2006_15:04:05.000000"), logType, "", logData)
	}

	if logFile != "" {
		// Check if log file exists, if not create it along with the path
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			os.MkdirAll(logFile, 0755)
		}

		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			tlog("error", fmt.Sprintf("%s", err))
		}
		defer f.Close()
		fmt.Fprintf(f, "%s [%s]% -s %s\n", time.Now().Format("02/01/2006_15:04:05.000000"), logType, "", logData)
	}
}

func sliceHas_string(s []string, n string) bool {
	for _, d := range s {
		if d == n {
			return true
		}
	}
	return false
}

func removeByIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func coordinator() {
	dmap := make(map[string]context.CancelFunc)
	executed := make([]string, 0)

	for {
		daemons := make([]string, 0)
		content, err := os.ReadFile(initFile)
		if err != nil {
			tlog("error", fmt.Sprintf("%s", err))
			os.Exit(3)
		}
		n := strings.Split(string(content), "\n")
		// Scan the initFile and load daemons and commands
		for _, line := range n {
			if strings.HasPrefix(line, "#") {
				continue
			}
			if line == "" {
				continue
			}

			params := strings.Split(line, ":")
			if len(params) < 2 {
				tlog("warning", fmt.Sprintf("Invalid config line [%s]. Skipping.", line))
				continue
			}

			cmdParam := ""
			for i := 1; i < len(params); i++ {
				if i == (len(params) - 1) {
					cmdParam += params[i]
				} else {
					cmdParam += params[i] + ":"
				}
			}

			switch params[0] {
			case "s":
				daemons = append(daemons, line)
				if !sliceHas_string(executed, line) {
					// For single run command there is no need to cancel context
					go executor(cmdParam, context.Background(), false)
					executed = append(executed, line)
				}

			case "d":
				daemons = append(daemons, cmdParam)
				// Check if daemon doesnt exist, then spawn
				_, exists := dmap[cmdParam]
				if !exists {
					ctx, cancel := context.WithCancel(context.Background())
					dmap[cmdParam] = cancel
					go executor(cmdParam, ctx, true)
				}
			}
		}

		// Go over dmap and and see if there are any daemons running that are no longer on the daemons list
		for daemon, c := range dmap {
			if !sliceHas_string(daemons, daemon) {
				tlog("info", fmt.Sprintf("[%s] => [STOPPING]", daemon))
				c()
				tlog("info", fmt.Sprintf("[%s] => [PURGING]", daemon))
				delete(dmap, daemon)
			}
		}

		for i, c := range executed {
			if !sliceHas_string(daemons, c) {
				tlog("info", fmt.Sprintf("[%s] => [PURGING]", c))
				executed = removeByIndex(executed, i)
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func executor(d string, ctx context.Context, daemon bool) {
	for {
		tlog("info", fmt.Sprintf("[%s] => [STARTED]", d))
		out, err := exec.CommandContext(ctx, "/bin/bash", "-u", "-o", "pipefail", "-c", d).Output()
		if err != nil {
			tlog("error", fmt.Sprintf("[%s] => [KILLED] | [%s] %s", d, err, out))
			if ctx.Err() != nil {
				if ctx.Err().Error() == "context canceled" {
					return
				}
			}
		}
		if !daemon {
			return
		}
		tlog("info", fmt.Sprintf("[%s] => [EXITED]", d))
		time.Sleep(3 * time.Second)
	}
}

func getSignal() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGCHLD)
	for {
		s := <-sigc
		switch s {
		case syscall.SIGCHLD:
			for {
				var (
					status syscall.WaitStatus
					usage  syscall.Rusage
				)
				pid, err := syscall.Wait4(-1, &status, syscall.WNOHANG, &usage)
				if pid < 1 {
					break
				}
				if err != nil {
					tlog("error", fmt.Sprintf("%s", err))
				}
				tlog("info", fmt.Sprintf("Reaping child PID: %v", pid))
			}
		}
	}
}

func handleArgs() {
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "update":
			// Determine OS and ARCH
			osRelease := runtime.GOOS
			arch := runtime.GOARCH

			// Build URL
			e := tuna.SelfUpdate(fmt.Sprintf("https://github.com/8ff/badger/releases/download/latest/badger.%s.%s", osRelease, arch))
			if e != nil {
				fmt.Println(e)
				os.Exit(1)
			}

			fmt.Println("Updated!")
			os.Exit(0)

		case "version", "-v", "--version":
			fmt.Fprintf(os.Stdout, "%s\n", "VERSION")
			os.Exit(0)

		case "help", "h", "-h", "--help":
			// Print help manually or call a helper function
			fmt.Println("Usage:")
			fmt.Println("  update    Selfupdate")
			fmt.Println("  version   Version")
			fmt.Println("  help      Usage")
			fmt.Println("  if        Initfile path. Takes [path] parameter")
			fmt.Println("  log       Logfile path. Takes [path] parameter")
			os.Exit(0)

		case "if":
			if i+1 < len(args) {
				initFile = args[i+1]
				i++
			}

		case "log":
			if i+1 < len(args) {
				logFile = args[i+1]
				i++
			}
		}
	}
}

func main() {
	// Defaults
	initFile = "/opt/runtime/if"
	logFile = "/opt/runtime/log/init.log"

	// Check if binary name is /sbin/init then set initFile to /initrc and log file to /init.log
	if strings.HasSuffix(os.Args[0], "/sbin/init") {
		initFile = "/initrc"
		logFile = "/init.log"
	}

	handleArgs()

	tlog("info", fmt.Sprintf("Initfile: %s", initFile))
	tlog("info", fmt.Sprintf("Logfile: %s", logFile))
	tlog("info", fmt.Sprintf("PID: %d", os.Getpid()))
	go getSignal()
	coordinator()
}
