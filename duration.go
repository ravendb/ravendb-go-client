package ravendb

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Duration is alias for time.Duration that serializes to JSON in a way that
// ravendb server understands
type Duration time.Duration

// MarshalJSON converts to JSON format (as a string)
func (d Duration) MarshalJSON() ([]byte, error) {
	dur := time.Duration(d)
	nDays := dur / (time.Hour * 24)
	dur = dur - (time.Hour * 24 * nDays)
	nHours := dur / time.Hour
	dur = dur - (time.Hour * nHours)
	nMins := dur / time.Minute
	dur = dur - (time.Minute * nMins)
	nSecs := dur / time.Second
	dur = dur - (time.Second * nSecs)
	var s string
	if nDays > 0 {
		// "5."
		s = fmt.Sprintf("%d.", nDays)
	}
	// "00:00:00"
	s += fmt.Sprintf("%02d:%02d:%02d", nHours, nMins, nSecs)

	millis := dur / time.Millisecond
	if millis > 0 {
		s += fmt.Sprintf(".%03d0000", millis)
	}
	s = `"` + s + `"`
	return []byte(s), nil
}

// UnmarshalJSON decodes from string
func (d *Duration) UnmarshalJSON(data []byte) error {
	s := string(data)
	errOut := fmt.Errorf("'%s' is not a valid JSON serializatioon of Duration", s)

	if len(s) < 8+2 || s[0] != '"' || s[len(s)-1] != '"' {
		// needs at least 00:00:00 + 2 * "
		return errOut
	}

	// string string quotes (") from the beginning and end of string
	s = s[1 : len(s)-1]
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return errOut
	}
	var nDays int
	var err error
	first := parts[0]
	// this can either be ${hour} or ${day}.${hour}
	parts2 := strings.Split(first, ".")
	if len(parts2) > 2 {
		return errOut
	}
	if len(parts2) == 2 {
		nDays, err = strconv.Atoi(parts2[0])
		if err != nil {
			return errOut
		}
		parts[0] = parts2[1]
	}
	// TODO: be more strict i.e. only allow 2-digit NN as a valid number
	nHours, err := strconv.Atoi(parts[0])
	if err != nil {
		return errOut
	}
	nMinutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return errOut
	}
	parts2 = strings.Split(parts[2], ".")
	if len(parts2) > 2 {
		return errOut
	}
	secsStr := parts[2]
	var nMS int
	if len(parts2) == 2 {
		secsStr = parts2[0]
		msStr := parts2[1]
		if len(msStr) > 3 {
			msStr = msStr[:3]
		}
		// ".1" is really ".100"
		for len(msStr) < 3 {
			msStr = msStr + "0"
		}
		nMS, err = strconv.Atoi(msStr)
		if err != nil {
			return errOut
		}
	}
	nSecs, err := strconv.Atoi(secsStr)
	if err != nil {
		return errOut
	}
	dur := time.Duration(nDays) * (time.Hour * 24)
	dur += time.Duration(nHours) * time.Hour
	dur += time.Duration(nMinutes) * time.Minute
	dur += time.Duration(nSecs) * time.Second
	dur += time.Duration(nMS) * time.Millisecond
	*d = Duration(dur)
	return nil
}
