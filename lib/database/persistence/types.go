package persistence

import "time"

type Configuration struct {
	SnapshotPath     string
	SnapshotInterval time.Duration
}
