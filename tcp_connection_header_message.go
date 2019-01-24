package ravendb

type OperationTypes = string

const (
	OperationNone           = "None"
	OperationDrop           = "Drop"
	OperationSubscription   = "Subscription"
	OperationReplicatoin    = "Replication"
	OperationCluster        = "Cluster"
	OperationHeartbeats     = "Heartbeats"
	OperationPing           = "Ping"
	OperationTestConnection = "TestConnection"
)

const (
	NUMBER_OR_RETRIES_FOR_SENDING_TCP_HEADER = 2
	SUBSCRIPTION_TCP_VERSION                 = 40
)

type TcpConnectionHeaderMessage struct {
	DatabaseName     string         `json:"DatabaseName"`
	SourceNodeTag    string         `json:"SourceNodeTag"`
	Operation        OperationTypes `json:"Operation"`
	OperationVersion int            `json:"OperationVersion"`
	Info             string         `json:"Info"`
}
