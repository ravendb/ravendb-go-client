package ravendb

import "time"

const (
	ISO8601TimeFormat = "2006-01-02T15:04:05.9999999Z"
)

func NetISO8601Utils_format(t time.Time) string {
	return t.Format(ISO8601TimeFormat)
}

func NetISO8601Utils_parse(s string) (time.Time, error) {
	return time.Parse(ISO8601TimeFormat, s)
}
