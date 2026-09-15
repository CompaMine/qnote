package main

import (
	"testing"
	"time"
)

func TestGetAndDeleteAtomic(t *testing.T) {
	s := NewStore()
	s.Put("abc", []byte("cipher"), time.Hour)

	got, ok := s.GetAndDelete("abc")
	if !ok || string(got) != "cipher" {
		t.Fatalf("expected cipher, got %q ok=%v", got, ok)
	}

	_, ok = s.GetAndDelete("abc")
	if ok {
		t.Fatal("second get should fail")
	}
}

func TestExpiredNote(t *testing.T) {
	s := NewStore()
	s.Put("x", []byte("c"), time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	_, ok := s.GetAndDelete("x")
	if ok {
		t.Fatal("expired note should not be returned")
	}
}

func TestCleanup(t *testing.T) {
	s := NewStore()
	s.Put("old", []byte("c"), time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	s.cleanup()
	if s.Len() != 0 {
		t.Fatalf("expected empty store, got %d", s.Len())
	}
}
