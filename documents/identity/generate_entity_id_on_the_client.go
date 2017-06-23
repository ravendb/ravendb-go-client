package identity

type OnClientIdGenerator struct{
	method func(interface{})string
}

func (generator OnClientIdGenerator) GenerateDocumentIdForStorage(entity interface{}) string{
	id := generator.GetOrGenerateDocumentId(entity)
	TrySetIdentity(entity, id)
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
