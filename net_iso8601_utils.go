package ravendb

import "time"

const (
	iso8601TimeFormat = "2006-01-02T15:04:05.9999999Z"
)

// TODO: needs to apply the same tweaks as in json marshaller for ravendb.Time

func NetISO8601UtilsFormat(t time.Time) string {
	return t.Format(iso8601TimeFormat)
}

func NetISO8601UtilsParse(s string) (time.Time, error) {
	return time.Parse(iso8601TimeFormat, s)
}
