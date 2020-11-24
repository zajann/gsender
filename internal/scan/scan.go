package scan

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
	log "github.com/zajann/easylog"
	"github.com/zajann/myip"
)

type Scanner struct {
	sourceDir  string
	renameDir  string
	fileRegexp *regexp.Regexp
	remoteHost string
	remotePort int
	remoteDir  string
	ftpUser    string
	ftpPass    string
	remove     bool
	mkdirByIP  bool
	interval   int
	realInf    string
}

func NewScanner(sourceDir string,
	renameDir string,
	reg string,
	remoteHost string,
	remotePort int,
	remoteDir string,
	ftpUser string,
	ftpPass string,
	remove bool,
	mkdirByIP bool,
	interval int,
	realInf string) (*Scanner, error) {
	s := &Scanner{
		sourceDir:  sourceDir,
		renameDir:  renameDir,
		remoteHost: remoteHost,
		remotePort: remotePort,
		remoteDir:  remoteDir,
		ftpUser:    ftpUser,
		ftpPass:    ftpPass,
		remove:     remove,
		mkdirByIP:  mkdirByIP,
		interval:   interval,
		realInf:    realInf,
	}
	var err error
	s.fileRegexp, err = regexp.Compile(reg)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Scanner) Start(wg *sync.WaitGroup, done <-chan bool) {
	defer wg.Done()
	t := s.interval
	for {
		var targets []string
		select {
		case <-done:
			return
		default:
			if t < s.interval {
				break
			} else {
				t = 0
			}
			log.Info("Start Scan Dir [%s]", s.sourceDir)
			files, err := ioutil.ReadDir(s.sourceDir)
			if err != nil {
				log.Error("Failed to scan directory[%s]: %s", s.sourceDir, err)
				break
			}
			for _, f := range files {
				if f.IsDir() {
					continue
				}
				if s.fileRegexp.MatchString(f.Name()) {
					targets = append(targets, fmt.Sprintf("%s/%s", s.sourceDir, f.Name()))
				}
			}
			if err := s.handleTargets(targets); err != nil {
				log.Error("Failed to handle target files: %s", err)
				break
			}
			log.Info("Finish Scan Dir [%s]", s.sourceDir)
		}
		time.Sleep(time.Second)
		t++
	}
}

func (s *Scanner) handleTargets(targets []string) error {
	if len(targets) == 0 {
		return nil
	}
	c, err := ftp.Dial(fmt.Sprintf("%s:%d", s.remoteHost, s.remotePort),
		ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return err
	}
	defer c.Quit()
	if err := c.Login(s.ftpUser, s.ftpPass); err != nil {
		return err
	}
	if err := c.ChangeDir(s.remoteDir); err != nil {
		return err
	}
	var curDir string
	curDir, err = c.CurrentDir()
	if err != nil {
		return err
	}
	if s.mkdirByIP {
		ip := myip.GetIPv4(s.realInf)
		if ip == "" {
			return errors.New("cannot get real ip")
		}
		entries, err := c.NameList(curDir)
		if err != nil {
			return err
		}
		var exist bool
		for _, e := range entries {
			if filepath.Base(e) == ip {
				exist = true
				break
			}
		}
		if !exist {
			if err := c.MakeDir(ip); err != nil {
				return err
			}
		}
		if err := c.ChangeDir(ip); err != nil {
			return err
		}
		curDir, err = c.CurrentDir()
		if err != nil {
			return err
		}
	}
	for _, t := range targets {
		f, err := os.Open(t)
		if err != nil {
			return err
		}
		fileName := filepath.Base(t)
		if err := c.Stor(fileName, f); err != nil {
			f.Close()
			return err
		}
		f.Close()
		if s.remove {
			if err := os.Remove(t); err != nil {
				return err
			}
		} else {
			if err := s.renameTarget(fileName); err != nil {
				return err
			}
		}
		log.Info("Success Send File: [%s] >>> [%s:%d%s/%s]",
			t,
			s.remoteHost,
			s.remotePort,
			curDir,
			fileName)
	}
	return nil
}

func (s *Scanner) renameTarget(f string) error {
	if _, err := os.Stat(s.renameDir); os.IsNotExist(err) {
		if err := os.MkdirAll(s.renameDir, os.ModePerm); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	oldPath := fmt.Sprintf("%s/%s", s.sourceDir, f)
	newPath := fmt.Sprintf("%s/%s.bak", s.renameDir, f)
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	return nil
}
