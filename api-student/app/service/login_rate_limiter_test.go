package service

import (
	"testing"
	"time"
)

func TestLoginRateLimiter_LockAfterMaxFails(t *testing.T) {
	l := NewLoginRateLimiter(6, time.Minute)
	defer l.Stop()

	key := "127.0.0.1|alice"

	for i := 1; i <= 5; i++ {
		locked, _ := l.RegisterFailure(key)
		if locked {
			t.Fatalf("expected belum lock pada percobaan %d", i)
		}
	}

	locked, retryAfter := l.RegisterFailure(key)
	if !locked {
		t.Fatal("expected lock pada percobaan ke-6")
	}
	if retryAfter <= 0 {
		t.Errorf("expected retryAfter > 0, dapat %d", retryAfter)
	}

	allowed, _ := l.Allow(key)
	if allowed {
		t.Error("expected Allow mengembalikan false saat terkunci")
	}
}

func TestLoginRateLimiter_SuccessResets(t *testing.T) {
	l := NewLoginRateLimiter(6, time.Minute)
	defer l.Stop()

	key := "127.0.0.1|budi"

	for i := 0; i < 5; i++ {
		_, _ = l.RegisterFailure(key)
	}

	l.RegisterSuccess(key)

	allowed, _ := l.Allow(key)
	if !allowed {
		t.Error("expected Allow true setelah RegisterSuccess")
	}

	if got := l.Size(); got != 0 {
		t.Errorf("expected Size 0 setelah reset, dapat %d", got)
	}
}

func TestLoginRateLimiter_Cleanup(t *testing.T) {
	l := NewLoginRateLimiter(6, 50*time.Millisecond)
	defer l.Stop()

	key := "127.0.0.1|charlie"
	for i := 0; i < 6; i++ {
		_, _ = l.RegisterFailure(key)
	}

	if l.Size() != 1 {
		t.Fatalf("expected Size 1 sebelum cleanup, dapat %d", l.Size())
	}

	time.Sleep(80 * time.Millisecond)
	l.Cleanup()

	if l.Size() != 0 {
		t.Errorf("expected Size 0 setelah cleanup, dapat %d", l.Size())
	}
}
