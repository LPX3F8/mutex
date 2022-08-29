# mutex
 🔐 in-memory mutex able to unlock with given token

## Install
```bash
go get -u github.com/LPX3F8/mutex
```
## Example
```go
import "github.com/LPX3F8/mutex"

func main() {
	tm := mutex.NewTokenMutex()
	tk := tm.Lock()         // return a string token, default a uuid
	tm.TryLock()            // return false, cause the mutex been locked
	tm.TryLockWithToken(tk) // return ture, cause the same token
	tm.LockWithToken(tk)    // return same token and will not block the process.
	tm.Unlock("fake" + tk)  // return false, cause token is wrong
	tm.Unlock(tk)           // return ture
}