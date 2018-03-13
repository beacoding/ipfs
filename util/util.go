package util

import (
	"testing"
	"time"
)

func SucceedsSoon(t *testing.T, f func() error) {
	timeout := time.After(time.Second * 2)
	c := make(chan error)
	go func() {
		for {
			select {
			case <-timeout:
				return
			default:
			}
			err := f()
			c <- err
			if err == nil {
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()
	var err error
	for {
		select {
		case err = <-c:
			if err == nil {
				return
			}
		case <-timeout:
			t.Fatalf("%+v", err)
		}
	}
}
