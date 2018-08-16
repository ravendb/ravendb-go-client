package ravendb

import (
	"fmt"
	"time"
)

// TODO: implementation could be improved
func TimeUtils_durationToTimeSpan(duration time.Duration) string {
	tm := int64(duration / time.Millisecond)
	millis := tm % 1000
	tm = tm / 1000 // seconds
	seconds := tm % 60
	tm = tm / 60 // in minutes
	minutes := tm % 60
	tm = tm / 60 // in hours
	hours := tm % 24
	tm = tm / 24 // in days
	days := tm

	s := ""

	if days > 0 {
		s += fmt.Sprintf("%d.", days)
	}
	s += fmt.Sprintf("%02d:", hours)
	s += fmt.Sprintf("%02d:", minutes)
	s += fmt.Sprintf("%02d", seconds)
	if millis > 0 {
		s += fmt.Sprintf(".%03d0000", millis)
	}
	return s
}
