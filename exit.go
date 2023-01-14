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

func Clean(fn func() error) {
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

func Done() {
	err := execHandles()
	if err != nil {
		log.Errorf("pid: %d exit failed: %s", pid, err)
	} else {
		log.Infof("pid: %d exit done", pid)
		os.Exit(0)
	}
}

func Signal() <-chan os.Signal {
	exit := make(chan os.Signal, 1)
	go func() {
		for {
			select {
			case s := <-signals:
				log.Infof("pid: %d received signal '%s' exiting...", pid, s)
				err := execHandles()
				if err != nil {
					log.Errorf("pid: %d exit failed: %s", pid, err)
				} else {
					log.Infof("pid: % exited", pid)
					close(signals)
					exit <- s
					return
				}
			}
		}
	}()
	return exit
}

func Wait() {
	for {
		select {
		case <-signals:
			log.Infof("pid: %d exiting...", pid)
			err := execHandles()
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

func execHandles() (err error) {
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
