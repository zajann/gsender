package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"

	log "github.com/zajann/easylog"
	"github.com/zajann/gsender/internal/config"
	"github.com/zajann/gsender/internal/scan"
	"github.com/zajann/process"
)

const (
	appVersion string = "0.1.0"
)

var (
	configFilePath string
)

func main() {
	parseFlag()

	// Load configuration file
	c, err := config.Load(configFilePath)
	if err != nil {
		fmt.Printf("[Error] Failed to load configuation file: %s\n", configFilePath)
		os.Exit(1)
	}

	// Check if process is already running
	running, err := process.IsRunning(fmt.Sprintf("%s/%s", c.PIDFilePath, c.PIDFileName))
	if err != nil {
		fmt.Printf("Failed to check process is already running: %s\n", err)
		os.Exit(1)
	}
	if running {
		fmt.Println("gSender is already running")
		os.Exit(0)
	}

	// Init log package
	if err := log.Init(
		log.SetFilePath(c.Log.FilePath),
		log.SetFileName(c.Log.FileName),
		log.SetLevel(log.LogLevel(c.Log.Level)),
		log.SetMaxSize(c.Log.MaxSize),
	); err != nil {
		fmt.Printf("[Error] Failed to init log: %s\n", err)
		os.Exit(1)
	}
	c.DumpToLog()

	done := make(chan bool, 1)

	// Handle signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Info("Received signal...[%d]", sig)
		close(done)
	}()

	// Start
	var scanners []*scan.Scanner
	for _, t := range c.Targets {
		scanner, err := scan.NewScanner(t.SourceDir,
			t.RenameDir,
			t.FileRegexp,
			t.RemoteHost,
			t.RemotePort,
			t.RemoteDir,
			t.FTPUser,
			t.FTPPasswd,
			t.Remove,
			t.MkdirByIP,
			t.Interval,
			c.App.RealInf)
		if err != nil {
			log.Fatal("Failed to create scanner: %s", err)
		}
		scanners = append(scanners, scanner)
	}
	wg := new(sync.WaitGroup)
	log.Info("gSender Start")
	for i := range scanners {
		wg.Add(1)
		go scanners[i].Start(wg, done)
	}
	wg.Wait()
	log.Info("gSender Shutdown")
}

func parseFlag() {
	var re string
	var target string

	t := flag.Bool("t", false, "test regular expression")
	flag.StringVar(&re, "regexp", "", "regular expression of file, required for test")
	flag.StringVar(&target, "target", "", "test file name, required for test")
	v := flag.Bool("v", false, "version")
	flag.StringVar(&configFilePath, "c", "", "configuration file path, required")

	flag.Parse()
	if *v {
		fmt.Printf("gsender Version...%s\n", appVersion)
		os.Exit(0)
	}
	if *t {
		testRegexp(re, target)
		os.Exit(0)
	}
	if configFilePath == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func testRegexp(r string, t string) {
	if r == "" || t == "" {
		flag.Usage()
		return
	}
	var result bool
	reg := regexp.MustCompile(r)
	if reg.MatchString(t) {
		result = true
	}
	fmt.Println(result)
}
