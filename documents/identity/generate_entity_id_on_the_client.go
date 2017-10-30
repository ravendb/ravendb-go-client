package identity

import (
)

type OnClientIdGenerator struct{
	method func(interface{})string
}

func NewOnClientIdGenerator()(*OnClientIdGenerator, error){
	return &OnClientIdGenerator{}, nil
}

func (generator OnClientIdGenerator) GenerateDocumentIdForStorage(entity interface{}) string{
	id := generator.GetOrGenerateDocumentId(entity)
	generator.TrySetIdentity(entity, id)
	return id
}

func (generator OnClientIdGenerator) GetOrGenerateDocumentId(entity interface{}) string{
	id, ok := LookupIdFromInstance(entity)
	if !ok{
		id = generator.method(entity)
	}
	//In C# code here is start_with_slash check. Look if we need it here too.
	return id
}

//func (generator OnClientIdGenerator) TrySetIdentity(entity interface{}, id string) error{
//	entityType := reflect.TypeOf(entity)
//	propertyFieldIdx, ok := data.LookupIdentityPropertyIdxByTag(entityType)
//	if ok{
//		return entityType.Field(propertyFieldIdx), true
//	}
//}
