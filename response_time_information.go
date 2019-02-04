package ravendb

import "time"

// ResponseTimeInformation describes timing information of server requests
type ResponseTimeInformation struct {
	totalServerDuration time.Duration
	totalClientDuration time.Duration

	durationBreakdown []ResponseTimeItem
}

func (i *ResponseTimeInformation) computeServerTotal() {
	var total time.Duration
	for _, rti := range i.durationBreakdown {
		total += rti.Duration
	}
	i.totalServerDuration = total
}

// ResponseTimeItem represents a duration for executing a given url
type ResponseTimeItem struct {
	URL      string
	Duration time.Duration
}
