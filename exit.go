package exit

import (
	"github.com/gozelle/logging"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	signals  = make(chan os.Signal, 1)
	cleanFns []func() error
	lock     = sync.Mutex{}
	pid      int
)

var log = logging.Logger("exit")

func init() {
	pid = os.Getpid()
	log.Infof("pid: %d start", pid)
	signal.Notify(signals,
		os.Interrupt,
		os.Kill,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
}

func Clean(fn func()) {
	lock.Lock()
	defer lock.Unlock()
	cleanFns = append(cleanFns, func() error {
		fn()
		return nil
	})
}

func Pid() int {
	return pid
}

func Exit() {
	signals <- os.Interrupt
}

func Signal() <-chan os.Signal {
	c := make(chan os.Signal, 1)
	go func() {
		for {
			select {
			case s := <-signals:
				log.Infof("pid: %d received signal: %s, exiting...", pid, s)
				err := clean()
				if err != nil {
					log.Errorf("pid: %d exit failed: %s", pid, err)
				} else {
					log.Infof("pid: %d exited", pid)
					close(signals)
					c <- s
					return
				}
			}
		}
	}()
	return c
}

func Wait() {
	for {
		select {
		case <-signals:
			log.Infof("pid: %d exiting...", pid)
			err := clean()
			if err != nil {
				log.Errorf("pid: %d exit failed : %s", pid, err)
			} else {
				log.Infof("pid: %d exited", pid)
				close(signals)
				os.Exit(0)
			}
		}
	}
}

func clean() (err error) {
	lock.Lock()
	defer lock.Unlock()
	for _, handler := range cleanFns {
		err = handler()
		if err != nil {
			return
		}
	}
	return
}
