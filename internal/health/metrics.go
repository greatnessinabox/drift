package health

import "time"

type Snapshot struct {
	Score     Score
	Timestamp time.Time
	CommitHash string
}
