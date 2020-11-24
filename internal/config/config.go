package config

import (
	"fmt"

	"github.com/jinzhu/configor"

	log "github.com/zajann/easylog"
)

type Config struct {
	Process
	Log
	Targets []Target
	App
}

type Process struct {
	PIDFilePath string `default:"./"`
	PIDFileName string `default:"gsender.pid"`
}

type Log struct {
	FilePath string `default:"../log"`
	FileName string `default:"gsender.log"`
	Level    int    `default:"0"`
	MaxSize  int    `default:"10"`
}

type Target struct {
	SourceDir  string
	RenameDir  string
	FileRegexp string
	RemoteHost string
	RemotePort int
	RemoteDir  string
	FTPUser    string
	FTPPasswd  string
	Remove     bool
	MkdirByIP  bool
	Interval   int
}

type App struct {
	RealInf string
}

func Load(filePath string) (*Config, error) {
	config := new(Config)

	if err := configor.Load(config, filePath); err != nil {
		return nil, err
	}
	return config, nil
}

func (c *Config) DumpToLog() {
	log.Info("############################################################################")
	log.Info("#")
	log.Info("# Settings")
	log.Info("#")
	log.Info("############################################################################")
	log.Info("%-35s : %s", "PID File Path", c.PIDFilePath)
	log.Info("%-35s : %s", "PID File Name", c.PIDFileName)
	log.Info("%-35s : %s", "Log File Path", c.Log.FilePath)
	log.Info("%-35s : %s", "Log File Name", c.Log.FileName)
	log.Info("%-35s : %d", "Log Level", c.Log.Level)
	log.Info("%-35s : %d", "Log File Max Size", c.Log.MaxSize)
	for i, t := range c.Targets {
		log.Info("%-35s : %s", fmt.Sprintf("Target[%d] Source Dir", i), t.SourceDir)
		log.Info("%-35s : %s", fmt.Sprintf("Target[%d] Rename Dir", i), t.RenameDir)
		log.Info("%-35s : %s", fmt.Sprintf("Target[%d] File Regexp", i), t.FileRegexp)
		log.Info("%-35s : %s", fmt.Sprintf("Target[%d] Remote Host", i), t.RemoteHost)
		log.Info("%-35s : %d", fmt.Sprintf("Target[%d] Remote Port", i), t.RemotePort)
		log.Info("%-35s : %s", fmt.Sprintf("Target[%d] Remote Dir", i), t.RemoteDir)
		log.Debug("%-35s : %s", fmt.Sprintf("Target[%d] FTP User", i), t.FTPUser)
		log.Debug("%-35s : %s", fmt.Sprintf("Target[%d] FTP Passwd", i), t.FTPPasswd)
		log.Info("%-35s : %v", fmt.Sprintf("Target[%d] Remove Source Files", i), t.Remove)
		log.Info("%-35s : %v", fmt.Sprintf("Target[%d] Make Remote Dir by IP", i), t.MkdirByIP)
		log.Info("%-35s : %d", fmt.Sprintf("Target[%d] Scan Interval(sec)", i), t.Interval)
	}
	log.Info("%-35s : %s", "Real Interface Name", c.App.RealInf)
	log.Info("############################################################################")
}
