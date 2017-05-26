package control

import (
	"encoding/json"
	"time"
)

// JobDuration is a simple structure used to embed a duration between 2 points in time.
// if t2 is zero time.Now() is used.
type JobDuration struct {
	t1, t2 *time.Time
}

// MarshalJSON makes the magic.
func (j JobDuration) MarshalJSON() ([]byte, error) {
	t2 := *j.t2
	if t2.IsZero() {
		t2 = time.Now()
	}

	s := t2.Sub(*j.t1).String()
	return json.Marshal(s)
}
