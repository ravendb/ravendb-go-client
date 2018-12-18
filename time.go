package ravendb

import (
	"encoding/json"
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
	// ravendb server only accepts 7 digits for fraction part but Go's
	// formatting might remove trailing zeros, producing 6 digits
	dotIdx := strings.LastIndexByte(s, '.')

	if dotIdx == -1 {
		s = s[:len(s)-1] // remove 'Z'
		s = s + ".0000000Z"
	} else {
		nToAdd := 9 - (len(s) - dotIdx) // 9: 7 + 1 for 'Z' and 1 for '.'
		if nToAdd > 0 {
			s = s[:len(s)-1] // remove 'Z'
			for ; nToAdd > 0; nToAdd-- {
				s = s + "0"
			}
			s = s + "Z"
		}
	}

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

// RoundToServerTime rounds t to the same precision as round-tripping
// to the server and back. Useful for comparing time.Time values for
// equality with values returned by the server
func RoundToServerTime(t time.Time) time.Time {
	st := Time(t)
	d, err := json.Marshal(st)
	must(err)
	var res Time
	err = json.Unmarshal(d, &res)
	must(err)
	return time.Time(res)
}
