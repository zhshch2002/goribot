package goribot

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	c := NewCacheManger()

	// test set
	if res := c.Set("1", 1*time.Hour, func() interface{} {
		t.Log("set cache 1")
		return 1
	}); res != 1 {
		t.Error("get wrong cache", res)
	}

	// test get
	if res := c.MustGet("1"); res != 1 {
		t.Error("get wrong cache", res)
	}

	// test GetAndSet
	if res := c.GetAndSet("2", 1*time.Second, func() interface{} {
		t.Log("set cache 2")
		return 2
	}); res != 2 {
		t.Error("get wrong cache", res)
	}
	if res := c.GetAndSet("2", 1*time.Second, func() interface{} {
		t.Log("set cache 2")
		return 2
	}); res != 2 {
		t.Error("get wrong cache", res)
	}

	time.Sleep(2 * time.Second)
	// test exp
	if res, ok := c.Get("2"); ok || res == 2 {
		t.Error("get wrong cache", res)
	}
	if res, ok := c.Get("2"); ok || res == 2 {
		t.Error("get wrong cache", res)
	}
}
