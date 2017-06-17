package data

import (
	"errors"
	"time"
	"fmt"
)

type BehaviorType uint8

const (
	LEADER_ONLY BehaviorType = iota
	LEADER_WITH_FAILOVER
	LEADER_WITH_FAILOVER_WHEN_REQUEST_TIME_SLA_THRESHOLD_IS_REACHED
	ROUND_ROBIN
	ROUND_ROBIN_WITH_FAILOVER_WHEN_REQUEST_TIME_SLA_THRESHOLD_IS_REACHED
	FASTEST_NODE
)

var ReadBehaviours = [...]string{
	"LeaderOnly",
	"LeaderWithFailover",
	"LeaderWithFailoverWhenRequestTimeSlaThresholdIsReached",
	"RoundRobin",
	"RoundRobinWithFailoverWhenRequestTimeSlaThresholdIsReached",
	"FastestNode",
}

var WriteBehaviours = [...]string{
	"LeaderOnly",
	"LeaderWithFailover",
}

type Behaviourer interface {
	getBehaviourName() string
}

type Behaviour struct {
	allowedBehaviours []string
	behaviorType BehaviorType
}

type ReadBehaviour struct {
	behaviour Behaviour
}

type WriteBehaviour struct {
	behaviour Behaviour
}

type DocumentConvention struct {
	MaxNumberOfRequestsPerSession, MaxIdsToCatch,
	Timeout, MaxLengthOfQueryUsingGetUrl uint
	DefaultUseOptimisticConcurrency bool
	IdentityPartsSeparator string
	JsonDefaultMethod func(obj interface{}) (interface{}, error)
}

func (b Behaviour) getBehaviourName() string{
	return b.allowedBehaviours[b.behaviorType]
}

func (b Behaviour) IsEmpty() bool{
	return len(b.allowedBehaviours) == 0 && b.behaviorType == 0
}

func (b ReadBehaviour) getBehaviourName() string{
	return b.behaviour.getBehaviourName()
}

func (b ReadBehaviour) IsEmpty() bool{
	return b.behaviour.IsEmpty()
}

func (b WriteBehaviour) getBehaviourName() string{
	return b.behaviour.getBehaviourName()
}

func (b WriteBehaviour) IsEmpty() bool{
	return b.behaviour.IsEmpty()
}

func NewBehaviour(allowedBehaviours []string, behaviourType BehaviorType) (*Behaviour, error){
	if int(behaviourType) >= len(allowedBehaviours){
		return nil, errors.New("data: Behaviour type out of range")
	}
	b := Behaviour{allowedBehaviours, behaviourType}
	return &b, nil
}

func NewReadBehaviour(behaviourType BehaviorType) (*ReadBehaviour, error){
	baseBehaviour, err := NewBehaviour(ReadBehaviours[:], behaviourType)
	if err != nil {
		return nil, err
	}
	b := ReadBehaviour{*baseBehaviour}
	return &b, nil
}

func NewWriteBehaviour(behaviourType BehaviorType) (*WriteBehaviour, error){
	baseBehaviour, err := NewBehaviour(WriteBehaviours[:], behaviourType)
	if err != nil {
		return nil, err
	}
	b := WriteBehaviour{*baseBehaviour}
	return &b, nil
}

func NewDocumentConvention() (*DocumentConvention, error){
	dc := DocumentConvention{
		30, 32,
		30, 1024 + 512,
		false,
		"/", jsonDefault,
	}
	return &dc, nil
}

func jsonDefault(obj interface{}) (interface{}, error){
	switch v := obj.(type) {
	default:
		return nil, errors.New(fmt.Sprintf("data: %#v is not JSON serializable (Try add a json default method to store convention)", obj))
	case nil:
		return nil, nil
	case time.Time:
		//TODO format datetime
		return v, nil
	case time.Duration:
		//TODO format datetime
		return v, nil
	}
}