package ravendb

import (
	"fmt"
	"strconv"
)

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
	PING_BASE_LINE                           = -1
	NONE_BASE_LINE                           = -1
	DROP_BASE_LINE                           = -2
	HEARTBEATS_BASE_LINE                     = 20
	SUBSCRIPTION_BASE_LINE                   = 40
	TEST_CONNECTION_BASE_LINE                = 50

	HEARTBEATS_TCP_VERSION      = HEARTBEATS_BASE_LINE
	SUBSCRIPTION_TCP_VERSION    = SUBSCRIPTION_BASE_LINE
	TEST_CONNECTION_TCP_VERSION = TEST_CONNECTION_BASE_LINE
)

type TcpConnectionHeaderMessage struct {
	DatabaseName     string         `json:"DatabaseName"`
	SourceNodeTag    string         `json:"SourceNodeTag"`
	Operation        OperationTypes `json:"Operation"`
	OperationVersion int            `json:"OperationVersion"`
	Info             string         `json:"Info"`
}

type PingFeatures struct {
	baseLine bool
}

func NewPingFeatures() *PingFeatures {
	return &PingFeatures{
		baseLine: true,
	}
}

type NoneFeatures struct {
	baseLine bool
}

func NewNoneFeatures() *NoneFeatures {
	return &NoneFeatures{
		baseLine: true,
	}
}

type DropFeatures struct {
	baseLine bool
}

func NewDropFeatures() *DropFeatures {
	return &DropFeatures{
		baseLine: true,
	}
}

type SubscriptionFeatures struct {
	baseLine bool
}

func NewSubscriptionFeatures() *SubscriptionFeatures {
	return &SubscriptionFeatures{
		baseLine: true,
	}
}

type HeartbeatsFeatures struct {
	baseLine bool
}

func NewHeartbeatsFeatures() *HeartbeatsFeatures {
	return &HeartbeatsFeatures{
		baseLine: true,
	}
}

type TestConnectionFeatures struct {
	baseLine bool
}

func NewTestConnectionFeatures() *TestConnectionFeatures {
	return &TestConnectionFeatures{
		baseLine: true,
	}
}

type ReplicationFeatures struct {
	baseLine           bool
	missingAttachments bool
}

func NewReplicationFeatures() *ReplicationFeatures {
	return &ReplicationFeatures{
		baseLine: true,
	}
}

type SupportedFeatures struct {
	protocolVersion int

	ping           *PingFeatures
	none           *NoneFeatures
	drop           *DropFeatures
	subscription   *SubscriptionFeatures
	heartbeats     *HeartbeatsFeatures
	testConnection *TestConnectionFeatures
}

func NewSupportedFeatures(version int) *SupportedFeatures {
	return &SupportedFeatures{
		protocolVersion: version,
	}
}

var (
	operationsToSupportedProtocolVersions = map[OperationTypes][]int{}
	supportedFeaturesByProtocol           = map[OperationTypes]map[int]*SupportedFeatures{}
)

func init() {
	operationsToSupportedProtocolVersions[OperationPing] = []int{PING_BASE_LINE}
	operationsToSupportedProtocolVersions[OperationNone] = []int{NONE_BASE_LINE}
	operationsToSupportedProtocolVersions[OperationDrop] = []int{DROP_BASE_LINE}
	operationsToSupportedProtocolVersions[OperationSubscription] = []int{SUBSCRIPTION_BASE_LINE}
	operationsToSupportedProtocolVersions[OperationHeartbeats] = []int{HEARTBEATS_BASE_LINE}
	operationsToSupportedProtocolVersions[OperationTestConnection] = []int{TEST_CONNECTION_BASE_LINE}

	pingFeaturesMap := map[int]*SupportedFeatures{}
	supportedFeaturesByProtocol[OperationPing] = pingFeaturesMap
	pingFeatures := NewSupportedFeatures(PING_BASE_LINE)
	pingFeatures.ping = NewPingFeatures()
	pingFeaturesMap[PING_BASE_LINE] = pingFeatures

	noneFeaturesMap := map[int]*SupportedFeatures{}
	supportedFeaturesByProtocol[OperationNone] = noneFeaturesMap
	noneFeatures := NewSupportedFeatures(NONE_BASE_LINE)
	noneFeatures.none = NewNoneFeatures()
	noneFeaturesMap[NONE_BASE_LINE] = noneFeatures

	dropFeaturesMap := map[int]*SupportedFeatures{}
	supportedFeaturesByProtocol[OperationDrop] = dropFeaturesMap
	dropFeatures := NewSupportedFeatures(DROP_BASE_LINE)
	dropFeatures.drop = NewDropFeatures()
	dropFeaturesMap[DROP_BASE_LINE] = dropFeatures

	subscriptionFeaturesMap := map[int]*SupportedFeatures{}
	supportedFeaturesByProtocol[OperationSubscription] = subscriptionFeaturesMap
	subscriptionFeatures := NewSupportedFeatures(SUBSCRIPTION_BASE_LINE)
	subscriptionFeatures.subscription = NewSubscriptionFeatures()
	subscriptionFeaturesMap[SUBSCRIPTION_BASE_LINE] = subscriptionFeatures

	heartbeatsFeaturesMap := map[int]*SupportedFeatures{}
	supportedFeaturesByProtocol[OperationHeartbeats] = heartbeatsFeaturesMap
	heartbeatsFeatures := NewSupportedFeatures(HEARTBEATS_BASE_LINE)
	heartbeatsFeatures.heartbeats = NewHeartbeatsFeatures()
	heartbeatsFeaturesMap[HEARTBEATS_BASE_LINE] = heartbeatsFeatures

	testConnectionFeaturesMap := map[int]*SupportedFeatures{}
	supportedFeaturesByProtocol[OperationTestConnection] = testConnectionFeaturesMap
	testConnectionFeatures := NewSupportedFeatures(TEST_CONNECTION_BASE_LINE)
	testConnectionFeatures.testConnection = NewTestConnectionFeatures()
	testConnectionFeaturesMap[TEST_CONNECTION_BASE_LINE] = testConnectionFeatures

}

var (
	// validate
	operations = []OperationTypes{
		OperationCluster,
		OperationDrop,
		OperationHeartbeats,
		OperationNone,
		OperationPing,
		OperationReplicatoin,
		OperationSubscription,
		OperationTestConnection,
	}
)

type SupportedStatus int

const (
	SupportedStatus_OUT_OF_RANGE SupportedStatus = iota
	SupportedStatus_NOT_SUPPORTED
	SupportedStatus_SUPPORTED
)

func operationVersionSupported(operationType OperationTypes, version int, currentRef *int) SupportedStatus {
	*currentRef = -1

	supportedProtocols := operationsToSupportedProtocolVersions[operationType]
	panicIf(len(supportedProtocols) == 0, "This is a bug. Probably you forgot to add '"+operationType+"' operation to the operationsToSupportedProtocolVersions map")

	for i := 0; i < len(supportedProtocols); i++ {
		*currentRef = supportedProtocols[i]
		if *currentRef == version {
			return SupportedStatus_SUPPORTED
		}

		if *currentRef < version {
			return SupportedStatus_NOT_SUPPORTED
		}
	}

	return SupportedStatus_OUT_OF_RANGE
}

func getOperationTcpVersion(operationType OperationTypes, index int) int {
	// we don't check the if the index go out of range, since this is expected and means that we don't have
	switch operationType {
	case OperationPing, OperationNone:
		return -1
	case OperationDrop:
		return -2
	case OperationSubscription, OperationReplicatoin, OperationCluster,
		OperationHeartbeats, OperationTestConnection:
		return operationsToSupportedProtocolVersions[operationType][index]
	default:
		panic(fmt.Sprintf("invalid operationType '%v'", operationType))
	}
}

func getSupportedFeaturesFor(typ OperationTypes, protocolVersion int) *SupportedFeatures {
	features := supportedFeaturesByProtocol[typ][protocolVersion]
	panicIf(features == nil, typ+" in protocol "+strconv.Itoa(protocolVersion)+" was not found in the features set")
	return features
}
