package service

import (
	"sync"
	"time"
)

type attempt struct {
	fails    int
	lockedAt time.Time
	lastSeen time.Time
}

type LoginRateLimiter struct {
	mu        sync.Mutex
	attempts  map[string]*attempt
	maxFails  int
	lockout   time.Duration
	stopCh    chan struct{}
	stopOnce  sync.Once
	wg        sync.WaitGroup
}

func NewLoginRateLimiter(maxFails int, lockout time.Duration) *LoginRateLimiter {
	return &LoginRateLimiter{
		attempts: make(map[string]*attempt),
		maxFails: maxFails,
		lockout:  lockout,
		stopCh:   make(chan struct{}),
	}
}

func (l *LoginRateLimiter) touch(a *attempt) {
	a.lastSeen = time.Now()
}

func (l *LoginRateLimiter) Allow(key string) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, ok := l.attempts[key]
	if !ok {
		return true, 0
	}

	if !a.lockedAt.IsZero() {
		remaining := time.Until(a.lockedAt)
		if remaining > 0 {
			l.touch(a)
			return false, int(remaining.Seconds()) + 1
		}
		delete(l.attempts, key)
		return true, 0
	}

	l.touch(a)
	return true, 0
}

func (l *LoginRateLimiter) RegisterFailure(key string) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, ok := l.attempts[key]
	if !ok {
		a = &attempt{}
		l.attempts[key] = a
	}

	a.fails++
	a.lastSeen = time.Now()
	if a.fails >= l.maxFails {
		a.lockedAt = time.Now().Add(l.lockout)
		retryAfter := int(l.lockout.Seconds())
		return true, retryAfter
	}

	return false, 0
}

func (l *LoginRateLimiter) RegisterSuccess(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}

func (l *LoginRateLimiter) Start(interval time.Duration) {
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-l.stopCh:
				return
			case <-t.C:
				l.Cleanup()
			}
		}
	}()
}

func (l *LoginRateLimiter) Stop() {
	l.stopOnce.Do(func() {
		close(l.stopCh)
		l.wg.Wait()
	})
}

func (l *LoginRateLimiter) Cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for k, a := range l.attempts {
		unlocked := a.lockedAt.IsZero() || now.After(a.lockedAt)
		idleLong := now.Sub(a.lastSeen) > l.lockout
		if unlocked && idleLong {
			delete(l.attempts, k)
		}
	}
}

func (l *LoginRateLimiter) Size() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.attempts)
}
