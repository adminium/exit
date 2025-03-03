package exit

import (
	"github.com/adminium/logger"
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

var log = logger.NewLogger("exit")

func init() {
	pid = os.Getpid()
	signal.Notify(
		signals,
		os.Interrupt,
		os.Kill,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
}

func PID() int {
	return pid
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

func Clean(fn func()) {
	lock.Lock()
	defer lock.Unlock()
	cleanFns = append(cleanFns, func() error {
		fn()
		return nil
	})
}

func Handle(fn func() error) {
	lock.Lock()
	defer lock.Unlock()
	cleanFns = append(cleanFns, fn)
}

func Pid() int {
	return pid
}

func Exit() {
	signals <- os.Interrupt
}

func exit() (ok bool) {
	log.Infof("exiting...")
	err := clean()
	if err != nil {
		log.Errorf("exit failed : %s", err)
		return
	}
	ok = true
	return
}

func Signal() chan struct{} {
	s := make(chan struct{})
	go func() {
		for {
			select {
			case <-signals:
				if exit() {
					s <- struct{}{}
					log.Infof("trigger exit signal")
					return
				}
			}
		}
	}()
	return s
}

func Wait() {
	for {
		select {
		case <-signals:
			if exit() {
				log.Infof("exit success")
				close(signals)
				os.Exit(0)
			}
		}
	}
}
