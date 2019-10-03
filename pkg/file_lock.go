package pkg

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// a simple implementation of file-lock between process, race condtion may occur but it's good enough for our purpose
type FileLock struct {
	path string
}

func NewFileLock(path string) FileLock {
	return FileLock{
		path: path,
	}
}

func (f FileLock) TryLock(timeout time.Duration) error {
	lockname := f.lockName()
	log.Debugf("acquiring file lock %s", lockname)
	start := time.Now()
	for {
		if start.Add(timeout).Before(time.Now()) {
			return fmt.Errorf("unable to acquire file lock %s after timemout of %s, resolve this by manully removing %s", lockname, PrettyDuration(timeout), lockname)
		}

		if f.Lock() {
			log.Debugf("file lock %s aquired", lockname)
			return nil
		}

		time.Sleep(1 * time.Second)
	}
}

// there can be a race condition, but it's good enough for CI/CD scenario
func (f FileLock) Lock() bool {
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

func (f FileLock) Unlock() {
	l := f.lockName()
	log.Debugf("release file lock %s", l)
	if err := os.Remove(l); err != nil {
		log.Debugf("err releasing file lock: %s", err.Error())
	}
}

func (f FileLock) lockName() string {
	return f.path + ".lock"
}
