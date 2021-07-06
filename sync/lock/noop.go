package lock

// NewNoopLocking 空locking
func NewNoopLocking() Locking {
	return &noopLocking{}
}

type noopLocking struct {
}

func (nl *noopLocking) Name() string {
	return "noop"
}

func (nl *noopLocking) Acquire(key string, opts *Options) (Locker, error) {
	return gNoopLocker, nil
}

var gNoopLocker = &noopLocker{}

type noopLocker struct {
}

func (nl *noopLocker) Lock() error {
	return nil
}

func (nl *noopLocker) Unlock() {
}
