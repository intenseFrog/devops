package pkg

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type FileLock interface {
	Lock() error
	Unlock()
}

func NewFileLock(path string, timeout time.Duration) FileLock {
	return newFileLockImplementation(path, timeout)
}

// a naive implementation of file-lock between process, race condtion may occur but it's good enough for our purpose
type fileLockImplementation struct {
	path    string
	timeout time.Duration
}

func newFileLockImplementation(path string, timeout time.Duration) *fileLockImplementation {
	return &fileLockImplementation{
		path:    path,
		timeout: timeout,
	}
}

func (f *fileLockImplementation) Lock() error {
	lockname := f.lockName()
	log.Debugf("acquiring file lock %s", lockname)
	expire := time.Now().Add(f.timeout)
	for {
		// try to acquire lock at least once
		if f.lock() {
			log.Debugf("file lock %s aquired", lockname)
			return nil
		}

		if time.Now().After(expire) {
			return fmt.Errorf("unable to acquire file lock %s after timemout of %s, wait till other process finsh working on %s or resolve this by manully removing %s", lockname, PrettyDuration(f.timeout), f.path, lockname)
		}

		time.Sleep(1 * time.Second)
	}
}

// there can be a race condition, but it's good enough for CI/CD scenario
func (f *fileLockImplementation) lock() bool {
	l := f.lockName()
	if _, err := os.Stat(l); err != nil {
		if !os.IsNotExist(err) {
			log.Debug(err.Error())
			return false
		}

		if _, err = os.Create(l); err != nil {
			log.Debug(err.Error())
			return false
		}

		return true
	}

	// lock busy
	return false
}

func (f *fileLockImplementation) Unlock() {
	l := f.lockName()
	log.Debugf("release file lock %s", l)
	if err := os.Remove(l); err != nil {
		log.Debugf("err releasing file lock: %s", err.Error())
	}
}

func (f *fileLockImplementation) lockName() string {
	return f.path + ".lock"
}
