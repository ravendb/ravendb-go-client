package ravendb

type tcpNegotiateParameters struct {
	operation          operationTypes
	version            int
	database           string
	sourceNodeTag      string
	destinationNodeTag string
	destinationUrl     string

	readResponseAndGetVersionCallback func(string) int
}
