package mutex

import (
	"runtime"
	"sync"
	"testing"
)

func TestNewTokenMutex(t *testing.T) {
	if n := runtime.SetMutexProfileFraction(1); n != 0 {
		t.Logf("got mutexrate %d expected 0", n)
	}
	defer runtime.SetMutexProfileFraction(0)
	l := NewTokenMutex()

	ok := false
	token := l.Lock()
	if _, ok = l.TryLock(); ok {
		t.Fatalf("TryLock succeeded with tokenLock locked")
	}
	if l.Unlock("faketoken") {
		t.Fatalf("Unlock succeeded with fake token")
	}
	nToken := l.LockWithToken(token)
	if nToken != token {
		t.Fatalf("LockWithToken token not equal")
	}
	if !l.Unlock(token) {
		t.Fatalf("Unlock failed with tokenLock locked")
	}
	if token, ok = l.TryLock(); !ok {
		t.Fatalf("TryLock failed with tokenLock unlocked")
	}
	if token, ok = l.TryLockWithToken(token); !ok {
		t.Fatalf("TryLockWithToken failed with tokenLock unlocked")
	}
	if !l.Unlock(token) {
		t.Fatalf("Unlocal failed with tokenLock locked")
	}

	wg := new(sync.WaitGroup)
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			HammerTokenMutex(l, 1000, wg)
		}()
	}
	wg.Wait()
}

func benchmarkTokenLock(b *testing.B, slack, work bool) {
	b.ReportAllocs()
	l := NewTokenMutex()
	if slack {
		b.SetParallelism(10)
	}
	b.RunParallel(func(pb *testing.PB) {
		foo := 0
		for pb.Next() {
			l.Unlock(l.Lock())
			if work {
				for i := 0; i < 100; i++ {
					foo *= 2
					foo /= 2
				}
			}
		}
		_ = foo
	})
}
func BenchmarkTokenMutex(b *testing.B) {
	benchmarkTokenLock(b, false, false)
}

func BenchmarkTokenMutexSlack(b *testing.B) {
	benchmarkTokenLock(b, true, false)
}

func BenchmarkTokenMutexWork(b *testing.B) {
	benchmarkTokenLock(b, false, true)
}

func BenchmarkTokenMutexWorkSlack(b *testing.B) {
	benchmarkTokenLock(b, true, true)
}

func BenchmarkTokenMutexNoSpin(b *testing.B) {
	var m = NewTokenMutex()
	var acc0, acc1 uint64
	var token string
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		c := make(chan bool)
		var data [4 << 10]uint64
		for i := 0; pb.Next(); i++ {
			if i%4 == 0 {
				token = m.Lock()
				acc0 -= 100
				acc1 += 100
				m.Unlock(token)
			} else {
				for i := 0; i < len(data); i += 4 {
					data[i]++
				}
				// Elaborate way to say runtime.Gosched
				// that does not put the goroutine onto global runq.
				go func() {
					c <- true
				}()
				<-c
			}
		}
	})
}

func BenchmarkTokenMutexSpin(b *testing.B) {
	var m = NewTokenMutex()
	var acc0, acc1 uint64
	var token string
	b.RunParallel(func(pb *testing.PB) {
		var data [16 << 10]uint64
		for i := 0; pb.Next(); i++ {
			token = m.Lock()
			acc0 -= 100
			acc1 += 100
			m.Unlock(token)
			for i := 0; i < len(data); i += 4 {
				data[i]++
			}
		}
	})
}

func HammerTokenMutex(m *TokenMutex, loops int, group *sync.WaitGroup) {
	for i := 0; i < loops; i++ {
		if i%3 == 0 {
			if token, ok := m.TryLock(); ok {
				m.Unlock(token)
			}
			continue
		}
		token := m.Lock()
		m.Unlock(token)
	}
	group.Done()
}
