package data

import (
	"errors"
	"time"
	"fmt"
	"reflect"
	"github.com/gedex/inflector"
	"strings"
	"unicode"
	"go/types"
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

const (
	COLLECTION = "@collection"
	METADATA_KEY = "@metadata"
	METADATA_ID = "@id"
	METADATA_ETAG = "@etag"
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

type Metadata struct{//todo make it dynamic
	Id string `json:"@id"`
	Collection string `json:"@collection"`
	ChangeVector string `json:"@change-vector"`
	LastModified time.Time `json:"@last-modified"`
	RavenGoType types.Type `json:"Raven-Go-Type"`
	RavenReadOnly bool `json:"Raven-Read-Only"`
}
///     The set of conventions used by the <see cref="DocumentStore" /> which allow the users to customize
///     the way the Raven client API behaves
type DocumentConvention struct {
	MaxNumberOfRequestsPerSession, MaxIdsToCatch,
	Timeout, MaxLengthOfQueryUsingGetUrl uint
	DefaultUseOptimisticConcurrency                 bool
	RaiseIfQueryPageSizeIsNotSet                    bool
	DisableTopologyUpdates                          bool
	IdentityPartsSeparator                          string
	JSONDefaultMethod                               func(obj interface{}) (interface{}, error)
	DocumentIdGenerator                             func(DBName string, entity interface{}) string
	registeredIdConventions                         map[string]func(DBName string, entity interface{}) string
	defaultCollectionNamesCache                     map[reflect.Type]string
	collectionNameFounder                           func(reflect.Type) (string, bool)
	TypeCollectionNameToDocumentIdPrefixTransformer func(string) string
}

func NewDocumentConvention() *DocumentConvention {
	dc := &DocumentConvention{
		MaxNumberOfRequestsPerSession:   30,
		MaxIdsToCatch:                   32,
		Timeout:                         30,
		MaxLengthOfQueryUsingGetUrl:     1024 + 512,
		DefaultUseOptimisticConcurrency: false,
		IdentityPartsSeparator:          "/",
		JSONDefaultMethod:               jsonDefault,
	}
	//todo: implement this
	//self.raise_if_query_page_size_is_not_set = kwargs.get("raise_if_query_page_size_is_not_set", False)

	//dc := &DocumentConvention{
	//	30,
	//	32,
	//	30,
	//	1024 + 512,
	//	false,
	//	"/",
	//	jsonDefault,
	//}
	return dc
}
// todo: совершенно непонимаю - что она должна делать
// имя обьекита по которому создается коллекция в базе в единичном числе, имя коллекции же в множдественном. эта функа преобразвоует
func (ref *DocumentConvention) defaultTransformPlural(name string) string{
	//Returns the plural form of a word if first parameter is greater than 1
	return "inflector.conditional_plural(2, name)"
}
func (ref *DocumentConvention) DefaultTransformTypeTagName(name string) string {
	count := 1
	for _, c := range []rune(name) {
		if c == unicode.ToUpper(c) {
			count ++
		}
	}
	// simple name, just lower case it
	if count <= 1 {
		return ref.defaultTransformPlural(strings.ToLower(name))
	}
	// multiple capital letters, so probably something that we want to preserve caps on.
	return ref.defaultTransformPlural(name)
	// @staticmethod
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

func LookupIdentityPropertyIdxByTag(entityType reflect.Type) (int, bool){
	for i := 0; i < entityType.NumField(); i++ {
		val := entityType.Field(i).Tag.Get("ravendb")
		if strings.HasSuffix(val, "id") || strings.Contains(val, "id,"){
			return i, true
		}
	}
	return -1, false
}

func (convention DocumentConvention) GenerateDocumentId(DBName string, entity interface{}) string{
	entityType := reflect.TypeOf(entity)
	registeredIdConvention, ok := convention.registeredIdConventions[entityType.String()]
	if ok{
		return registeredIdConvention(DBName, entity)
	}
	return convention.DocumentIdGenerator(DBName, entity)
}

func (convention DocumentConvention) GenerateDocumentIdAsync(DBName string, entity interface{}) <-chan string{
	out := make(chan string, 1)
	go func(){
		out <- convention.GenerateDocumentId(DBName, entity)
		close(out)
	}()
	return out
}

func (convention DocumentConvention) GetCollectionName(entity interface{}) string{
	if entity == nil{
		return ""
	}
	entityType := reflect.TypeOf(entity)
	result, ok := convention.collectionNameFounder(entityType)
	if !ok{
		result = convention.getDefaultCollectionName(entityType)
	}

	return result
}

func (convention DocumentConvention) getDefaultCollectionName(t reflect.Type) string{
	if _, ok := convention.defaultCollectionNamesCache[t]; !ok{
		convention.defaultCollectionNamesCache[t] = inflector.Pluralize(t.Name())
	}
	return convention.defaultCollectionNamesCache[t]
}

func (convention DocumentConvention) TryGetTypeFromMetadata(metadata Metadata) (types.Type, bool){
	if metadata.RavenGoType != nil {
		return metadata.RavenGoType, true
	}
	return nil, false
}
