package ravendb

import (
	"strings"
	"time"
)

const (
	// time format returned by the server which looks like:
	// 2018-05-08T05:20:31.5233900Z or
	// 2018-07-29T23:50:57.9998240 or
	// 2018-08-16T13:56:59.355664-07:00
	timeFormat  = "2006-01-02T15:04:05.9999999Z"
	timeFormat2 = "2006-01-02T15:04:05.9999999"
	timeFormat3 = "2006-01-02T15:04:05.9999999-07:00"
)

// Time is an alias for time.Time that serializes/deserializes in ways
// compatible with Ravendb server
type Time time.Time

func (t Time) MarshalJSON() ([]byte, error) {
	s := time.Time(t).Format(timeFormat)
	return []byte(`"` + s + `"`), nil
}

func (t *Time) UnmarshalJSON(d []byte) error {
	s := string(d)
	s = strings.TrimLeft(s, `"`)
	s = strings.TrimRight(s, `"`)

	if s == "null" {
		return nil
	}

	tt, err := time.Parse(timeFormat, s)
	if err != nil {
		tt, err = time.Parse(timeFormat2, s)
		if err != nil {
			tt, err = time.Parse(timeFormat3, s)
		}
	}
	if err != nil {
		// TODO: for now make it a fatal error to catch bugs early
		must(err)
		return err
	}

	*t = Time(tt)
	return nil
}

func (t *Time) toTime() time.Time {
	// for convenience make it work on nil pointer
	if t == nil {
		return time.Time{}
	}
	return time.Time(*t)
}

func (t *Time) toTimePtr() *time.Time {
	if t == nil {
		return nil
	}
	res := time.Time(*t)
	return &res
}
