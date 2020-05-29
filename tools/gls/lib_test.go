package gls

import (
	"testing"
	"time"
)

func TestGLS(t *testing.T) {

	_, ok := Get("abc")
	if ok {
		t.Error("must by not exists")
	}
	Set("abc", 1)

	_, ok = Get("abc")
	if !ok {
		t.Error("must  exists")
	}
	go (func() {
		_, ok = Get("abc")
		if ok {
			t.Error("must  by not exists")
		}
	})()

	Shutdown()
	_, ok = Get("abc")
	if ok {
		t.Error("must by not exists")
	}

	time.Sleep(1 * time.Second)

}
