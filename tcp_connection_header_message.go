package ravendb

import (
	"fmt"
	"strconv"
)

type operationTypes = string

const (
	operationNone           = "None"
	operationDrop           = "Drop"
	operationSubscription   = "Subscription"
	operationReplication    = "Replication"
	operationCluster        = "Cluster"
	operationHeartbeats     = "Heartbeats"
	operationPing           = "Ping"
	operationTestConnection = "TestConnection"
)

const (
	numberOfRetriesForSendingTCPHeader = 2
	pingBaseLine                       = -1
	noneBaseLine                       = -1
	dropBaseLine                       = -2
	hearthbeatsBaseLine                = 20
	subscriptionBaseLine               = 40
	testConnectionBaseLine             = 50

	heartbeatsTCPVersion     = hearthbeatsBaseLine
	subscriptionTCPVersion   = subscriptionBaseLine
	testConnectionTCPVersion = testConnectionBaseLine
)

type tcpConnectionHeaderMessage struct {
	DatabaseName     string         `json:"DatabaseName"`
	SourceNodeTag    string         `json:"SourceNodeTag"`
	Operation        operationTypes `json:"Operation"`
	OperationVersion int            `json:"OperationVersion"`
	Info             string         `json:"Info"`
}

type pingFeatures struct {
	baseLine bool
}

func newPingFeatures() *pingFeatures {
	return &pingFeatures{
		baseLine: true,
	}
}

type noneFeatures struct {
	baseLine bool
}

func newNoneFeatures() *noneFeatures {
	return &noneFeatures{
		baseLine: true,
	}
}

type dropFeatures struct {
	baseLine bool
}

func newDropFeatures() *dropFeatures {
	return &dropFeatures{
		baseLine: true,
	}
}

type subscriptionFeatures struct {
	baseLine bool
}

func newSubscriptionFeatures() *subscriptionFeatures {
	return &subscriptionFeatures{
		baseLine: true,
	}
}

type heartbeatsFeatures struct {
	baseLine bool
}

func newHeartbeatsFeatures() *heartbeatsFeatures {
	return &heartbeatsFeatures{
		baseLine: true,
	}
}

type testConnectionFeatures struct {
	baseLine bool
}

func newTestConnectionFeatures() *testConnectionFeatures {
	return &testConnectionFeatures{
		baseLine: true,
	}
}

type replicationFeatures struct {
	baseLine           bool
	missingAttachments bool
}

func newReplicationFeatures() *replicationFeatures {
	return &replicationFeatures{
		baseLine: true,
	}
}

type supportedFeatures struct {
	protocolVersion int

	ping           *pingFeatures
	none           *noneFeatures
	drop           *dropFeatures
	subscription   *subscriptionFeatures
	heartbeats     *heartbeatsFeatures
	testConnection *testConnectionFeatures
}

func newSupportedFeatures(version int) *supportedFeatures {
	return &supportedFeatures{
		protocolVersion: version,
	}
}

var (
	operationsToSupportedProtocolVersions = map[operationTypes][]int{}
	supportedFeaturesByProtocol           = map[operationTypes]map[int]*supportedFeatures{}
)

func init() {
	operationsToSupportedProtocolVersions[operationPing] = []int{pingBaseLine}
	operationsToSupportedProtocolVersions[operationNone] = []int{noneBaseLine}
	operationsToSupportedProtocolVersions[operationDrop] = []int{dropBaseLine}
	operationsToSupportedProtocolVersions[operationSubscription] = []int{subscriptionBaseLine}
	operationsToSupportedProtocolVersions[operationHeartbeats] = []int{hearthbeatsBaseLine}
	operationsToSupportedProtocolVersions[operationTestConnection] = []int{testConnectionBaseLine}

	pingFeaturesMap := map[int]*supportedFeatures{}
	supportedFeaturesByProtocol[operationPing] = pingFeaturesMap
	pingFeatures := newSupportedFeatures(pingBaseLine)
	pingFeatures.ping = newPingFeatures()
	pingFeaturesMap[pingBaseLine] = pingFeatures

	noneFeaturesMap := map[int]*supportedFeatures{}
	supportedFeaturesByProtocol[operationNone] = noneFeaturesMap
	noneFeatures := newSupportedFeatures(noneBaseLine)
	noneFeatures.none = newNoneFeatures()
	noneFeaturesMap[noneBaseLine] = noneFeatures

	dropFeaturesMap := map[int]*supportedFeatures{}
	supportedFeaturesByProtocol[operationDrop] = dropFeaturesMap
	dropFeatures := newSupportedFeatures(dropBaseLine)
	dropFeatures.drop = newDropFeatures()
	dropFeaturesMap[dropBaseLine] = dropFeatures

	subscriptionFeaturesMap := map[int]*supportedFeatures{}
	supportedFeaturesByProtocol[operationSubscription] = subscriptionFeaturesMap
	subscriptionFeatures := newSupportedFeatures(subscriptionBaseLine)
	subscriptionFeatures.subscription = newSubscriptionFeatures()
	subscriptionFeaturesMap[subscriptionBaseLine] = subscriptionFeatures

	heartbeatsFeaturesMap := map[int]*supportedFeatures{}
	supportedFeaturesByProtocol[operationHeartbeats] = heartbeatsFeaturesMap
	heartbeatsFeatures := newSupportedFeatures(hearthbeatsBaseLine)
	heartbeatsFeatures.heartbeats = newHeartbeatsFeatures()
	heartbeatsFeaturesMap[hearthbeatsBaseLine] = heartbeatsFeatures

	testConnectionFeaturesMap := map[int]*supportedFeatures{}
	supportedFeaturesByProtocol[operationTestConnection] = testConnectionFeaturesMap
	testConnectionFeatures := newSupportedFeatures(testConnectionBaseLine)
	testConnectionFeatures.testConnection = newTestConnectionFeatures()
	testConnectionFeaturesMap[testConnectionBaseLine] = testConnectionFeatures

}

var (
	// validate
	operations = []operationTypes{
		operationCluster,
		operationDrop,
		operationHeartbeats,
		operationNone,
		operationPing,
		operationReplication,
		operationSubscription,
		operationTestConnection,
	}
)

type supportedStatus int

const (
	supportedStatus_OUT_OF_RANGE supportedStatus = iota
	supportedStatus_NOT_SUPPORTED
	supportedStatus_SUPPORTED
)

func operationVersionSupported(operationType operationTypes, version int, currentRef *int) supportedStatus {
	*currentRef = -1

	supportedProtocols := operationsToSupportedProtocolVersions[operationType]
	panicIf(len(supportedProtocols) == 0, "This is a bug. Probably you forgot to add '"+operationType+"' operation to the operationsToSupportedProtocolVersions map")

	for i := 0; i < len(supportedProtocols); i++ {
		*currentRef = supportedProtocols[i]
		if *currentRef == version {
			return supportedStatus_SUPPORTED
		}

		if *currentRef < version {
			return supportedStatus_NOT_SUPPORTED
		}
	}

	return supportedStatus_OUT_OF_RANGE
}

func getOperationTcpVersion(operationType operationTypes, index int) int {
	// we don't check the if the index go out of range, since this is expected and means that we don't have
	switch operationType {
	case operationPing, operationNone:
		return -1
	case operationDrop:
		return -2
	case operationSubscription, operationReplication, operationCluster,
		operationHeartbeats, operationTestConnection:
		return operationsToSupportedProtocolVersions[operationType][index]
	default:
		panic(fmt.Sprintf("invalid operationType '%v'", operationType))
	}
}

func getSupportedFeaturesFor(typ operationTypes, protocolVersion int) *supportedFeatures {
	features := supportedFeaturesByProtocol[typ][protocolVersion]
	panicIf(features == nil, typ+" in protocol "+strconv.Itoa(protocolVersion)+" was not found in the features set")
	return features
}
