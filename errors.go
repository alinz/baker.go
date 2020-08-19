package baker

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrWatcherClosed = errors.New("watcher closed")
	ErrPingerClosed  = errors.New("pinger closed")
)
