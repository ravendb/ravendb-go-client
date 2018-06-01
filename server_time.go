package ravendb

import "time"

const (
	// time format returned by the server which looks like:
	// 2018-05-08T05:20:31.5233900Z
	serverTimeFormat = "2006-01-02T15:04:05.999999999Z"
)

type ServerTime time.Time

func (t ServerTime) MarshalJSON() ([]byte, error) {
	s := time.Time(t).Format(serverTimeFormat)
	return []byte(s), nil
}

func (t *ServerTime) UnmarshalJSON(d []byte) error {
	tt, err := time.Parse(serverTimeFormat, string(d))
	if err != nil {
		return err
	}
	*t = ServerTime(tt)
	return nil
}
