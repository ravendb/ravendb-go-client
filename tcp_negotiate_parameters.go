package ravendb

type TcpNegotiateParameters struct {
	operation          OperationTypes
	version            int
	database           string
	sourceNodeTag      string
	destinationNodeTag string
	destinationUrl     string

	readResponseAndGetVersionCallback func(string) int
}
