package ravendb

import (
	"bytes"
	"strings"
)

// CompareExchangeValue represents value for compare exchange
type CompareExchangeValue struct {
	Key          string
	Index        int64
	Value        interface{}
	ChangeVector *string
	Metadata     *MetadataAsDictionary
}

type ICompareExchangeValue interface {
	GetKey() string
	GetIndex() int64
	SetIndex(int64)
	GetValue() interface{}
	GetMetadata() *MetadataAsDictionary
	HasMetadata() bool
}

func (cev *CompareExchangeValue) getPropertyFromValue(key string) interface{} {
	if cev.Value == nil {
		return nil
	}

	object, isMap := cev.Value.(map[string]interface{})
	if isMap == false {
		return nil
	}

	keyValue, exists := object[key]
	if exists == false {
		return nil
	}

	return keyValue
}

func (cev *CompareExchangeValue) hasChanged(other *CompareExchangeValue) (bool, error) {
	if cev == other {
		return false, nil
	} // ptr equals

	if strings.EqualFold(cev.GetKey(), other.GetKey()) == false {
		return false, newIllegalArgumentError("Keys do not match. Expected: " + cev.Key + " but was " + other.Key)
	}

	if cev.Index != other.Index {
		return true, nil
	}

	first, err := jsonMarshal(cev.Value)
	if err != nil {
		return false, err
	}
	second, err := jsonMarshal(other.Value)
	if err != nil {
		return false, err
	}

	return bytes.Equal(first, second) == false, nil //compare
}

func (cev *CompareExchangeValue) GetKey() string                     { return cev.Key }
func (cev *CompareExchangeValue) GetIndex() int64                    { return cev.Index }
func (cev *CompareExchangeValue) SetIndex(index int64)               { cev.Index = index }
func (cev *CompareExchangeValue) GetValue() interface{}              { return cev.Value }
func (cev *CompareExchangeValue) GetMetadata() *MetadataAsDictionary { return cev.Metadata }
func (cev *CompareExchangeValue) HasMetadata() bool                  { return cev.Metadata != nil }

// NewCompareExchangeValue returns new CompareExchangeValue
func NewCompareExchangeValue(key string, index int64, value interface{}) *CompareExchangeValue {
	return NewCompareExchangeValueBase(key, index, value, nil, nil)
}

func NewCompareExchangeValueBase(key string, index int64, value interface{}, changeVector *string, dictionary *MetadataAsDictionary) *CompareExchangeValue {
	return &CompareExchangeValue{
		Key:          key,
		Index:        index,
		Value:        value,
		ChangeVector: changeVector,
		Metadata:     dictionary,
	}
}
