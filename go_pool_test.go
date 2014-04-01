package gopool

import "testing"
import "time"

func TestInstantiate(t *testing.T) {
	var gp GoPool //Explicit type here to make New returns a GoPool
	var err error
	gp, err = New(100)
	if err != nil {
		t.Error(err)
	}
	if gp == nil {
		t.Error("gp was nil")
	}
}

func TestInstantiateMax0(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Error("New accepted 0 pool size")
	}
}

func TestGoPoolBasic(t *testing.T) {
	gp, err := New(1)
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan bool)
	timer := time.After(100 * time.Millisecond)
	gp.Go(func() { ch <- true })
	select {
	case _ = <-ch:
	case _ = <-timer:
		t.Error("Timed out")
	}
}

func TestGoPoolConc(t *testing.T) {
	gp, err := New(1)
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan bool)
	//The lines below will timeout if gp.Go doesn't run concurrently
	gp.Go(func() {
		timer := time.After(100 * time.Millisecond)
		select {
		case _ = <-ch:
		case _ = <-timer:
			t.Error("Timed out")
		}
	})
	ch <- true
}
