package exit

import (
	"testing"
	"time"
)

func TestSignal(t *testing.T) {
	
	s := Signal()
	
	go func() {
		time.Sleep(2 * time.Second)
		Exit()
	}()
	
	select {
	case v := <-s:
		t.Log("退出", v)
	}
}
