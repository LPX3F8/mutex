package mutex

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

var _zeroString string

type TokenMutex struct {
	token string
	mu    *sync.Mutex
}

func NewTokenMutex() *TokenMutex {
	return &TokenMutex{
		token: _zeroString,
		mu:    new(sync.Mutex),
	}
}

// Lock will lock the mutex and return a token until mutex be locked
// This mutex can only be unlocked with this token
func (m *TokenMutex) Lock() string {
	return m.LockWithToken(_zeroString)
}

// LockWithToken will lock the mutex and return a token until mutex be locked
// This mutex can only be unlocked with this token
// The user can specify a token. If the lock is held by this token,
// it will be deemed to have been locked successfully.
// Note: When using it, it should be clear the whole process.
func (m *TokenMutex) LockWithToken(token string) string {
	var t string
	var s bool
	for t, s = m.TryLockWithToken(token); !s; t, s = m.TryLockWithToken(token) {
		time.Sleep(time.Nanosecond)
	}
	return t
}

// TryLock will try to lock the mutex and return lock result
// If lock is successful, will return token and true
// This mutex can only be unlocked with this token
func (m *TokenMutex) TryLock() (string, bool) {
	return m.TryLockWithToken(_zeroString)
}

// TryLockWithToken will try to lock the mutex and return lock result
// The user can specify a token. If the lock is held by this token,
// it will be deemed to have been locked successfully.
// If lock is successful, will return token and true
// This mutex can only be unlocked with this token
// Note: When using it, it should be clear the whole process.
func (m *TokenMutex) TryLockWithToken(token string) (string, bool) {
	if !m.mu.TryLock() {
		return _zeroString, false
	}
	defer m.mu.Unlock()
	if t := m.token; t != _zeroString && t != token {
		return _zeroString, false
	}
	if token == _zeroString {
		token = uuid.NewString()
	}
	m.token = token
	return token, true
}

func (m *TokenMutex) Unlock(token string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.token == token || m.token == _zeroString {
		m.token = _zeroString
		return true
	}
	return false
}
