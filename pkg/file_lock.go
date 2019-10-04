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
	return newFileLockImpl(path, timeout)
}

// a naive implementation of file-lock between process, race condtion may occur but it's good enough for our purpose
type fileLockImpl struct {
	path    string
	timeout time.Duration
}

func newFileLockImpl(path string, timeout time.Duration) *fileLockImpl {
	return &fileLockImpl{
		path:    path,
		timeout: timeout,
	}
}

func (f *fileLockImpl) Lock() error {
	l := f.lockName()
	log.Debugf("acquiring file lock %s", l)
	expire := time.Now().Add(f.timeout)
	for {
		// try to acquire lock at least once
		if f.lock() {
			log.Debugf("file lock %s aquired", l)
			return nil
		}

		if time.Now().After(expire) {
			return fmt.Errorf("unable to acquire file lock %s after timemout of %s, wait till other process finish working on %s or resolve this by manully removing %s", l, PrettyDuration(f.timeout), f.path, l)
		}

		time.Sleep(1 * time.Second)
	}
}

// there can be a race condition, but it's good enough for CI/CD scenario
func (f *fileLockImpl) lock() bool {
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

func (f *fileLockImpl) Unlock() {
	l := f.lockName()
	log.Debugf("release file lock %s", l)
	if err := os.Remove(l); err != nil {
		log.Debugf("err releasing file lock: %s", err.Error())
	}
}

func (f *fileLockImpl) lockName() string {
	return f.path + ".lock"
}
