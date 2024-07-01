package shared

import "sync"

type resources struct {
	UserLocks sync.Map
}

var Resources resources

func (r *resources) GetUserLock(userID string) *sync.Mutex {
	lock, _ := r.UserLocks.LoadOrStore(userID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}
