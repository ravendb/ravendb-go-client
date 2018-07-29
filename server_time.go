package ravendb

import (
	"strings"
	"time"
)

const (
	// time format returned by the server which looks like:
	// 2018-05-08T05:20:31.5233900Z
	serverTimeFormat = "2006-01-02T15:04:05.9999999Z"
)

type ServerTime time.Time

func (t ServerTime) MarshalJSON() ([]byte, error) {
	s := time.Time(t).Format(serverTimeFormat)
	return []byte(s), nil
}

func (t *ServerTime) UnmarshalJSON(d []byte) error {
	s := string(d)
	s = strings.TrimLeft(s, `"`)
	s = strings.TrimRight(s, `"`)

	if s == "null" {
		return nil
	}

	tt, err := time.Parse(serverTimeFormat, s)
	if err != nil {
		// server sometimes returns this value, which is missing
		// "Z" at the end so doesn't parse as serverTimeFormat
		if s == "0001-01-01T00:00:00.0000000" {
			tt = time.Time{}
		} else {
			// TODO: for now make it a fatal error to catch bugs early
			must(err)
			return err
		}
	}
	*t = ServerTime(tt)
	return nil
}

func (t *ServerTime) toTime() time.Time {
	// for convenience make it work on nil pointer
	if t == nil {
		return time.Time{}
	}
	return time.Time(*t)
}

func (t *ServerTime) toTimePtr() *time.Time {
	if t == nil {
		return nil
	}
	res := time.Time(*t)
	return &res
}

func serverTimePtrToTimePtr(t *ServerTime) *time.Time {
	return t.toTimePtr()
}
