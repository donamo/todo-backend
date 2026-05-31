package config

import (
	"testing"
	"time"
)

func TestDurationFallback(t *testing.T) {
	t.Setenv("TODO_BACKEND_TEST_EMPTY", "")
	got, err := Duration("TODO_BACKEND_TEST_EMPTY", 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got != 3*time.Second {
		t.Fatalf("duration = %s", got)
	}
}

func TestDurationInvalid(t *testing.T) {
	t.Setenv("TODO_BACKEND_TEST_DURATION", "soon")
	_, err := Duration("TODO_BACKEND_TEST_DURATION", time.Second)
	if err == nil {
		t.Fatal("expected error")
	}
}
