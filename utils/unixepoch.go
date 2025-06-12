package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// https://ikso.us/posts/unmarshal-timestamp-as-time/

// UnixTime is our magic type
type UnixTime struct {
	time.Time
}

// UnmarshalJSON is the method that satisfies the Unmarshaller interface
// Note that it uses a pointer receiver. It needs this because it will be modifying the embedded time.Time instance
func (u *UnixTime) UnmarshalJSON(b []byte) error {
	var timestamp int64
	err := json.Unmarshal(b, &timestamp)
	if err != nil {
		return err
	}
	u.Time = time.Unix(timestamp, 0)
	return nil
}

// MarshalJSON turns our time.Time back into an int
func (u UnixTime) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, "%d", (u.Time.Unix())), nil
}
