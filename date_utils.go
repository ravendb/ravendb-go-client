package ravendb

import "time"

func DateUtils_addHours(t time.Time, nHours int) time.Time {
	return t.Add(time.Hour * time.Duration(nHours))
}

func DateUtils_addDays(t time.Time, nDays int) time.Time {
	return DateUtils_addHours(t, nDays*24)
}

func DateUtils_addMinutes(t time.Time, nMinutes int) time.Time {
	return t.Add(time.Minute * time.Duration(nMinutes))
}
