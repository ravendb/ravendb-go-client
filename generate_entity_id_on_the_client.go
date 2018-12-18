package ravendb

type GenIDFunc func(interface{}) string

type GenerateEntityIDOnTheClient struct {
	_conventions *DocumentConventions
	_generateId  GenIDFunc
}

func NewGenerateEntityIDOnTheClient(conventions *DocumentConventions, generateId GenIDFunc) *GenerateEntityIDOnTheClient {
	return &GenerateEntityIDOnTheClient{
		_conventions: conventions,
		_generateId:  generateId,
	}
}

// Attempts to get the document key from an instance
func (g *GenerateEntityIDOnTheClient) tryGetIDFromInstance(entity interface{}) (string, bool) {
	panicIf(entity == nil, "Entity cannot be null")
	return tryGetIDFromInstance(entity)
}

// Tries to get the identity.
func (g *GenerateEntityIDOnTheClient) getOrGenerateDocumentID(entity interface{}) string {
	id, ok := g.tryGetIDFromInstance(entity)
	if !ok || id == "" {
		id = g._generateId(entity)
	}

	/* TODO:
	        if (id != null && id.startsWith("/")) {
	            throw new IllegalStateException("Cannot use value '" + id + "' as a document id because it begins with a '/'");
			}
	*/
	return id
}

func (g *GenerateEntityIDOnTheClient) generateDocumentKeyForStorage(entity interface{}) string {
	id := g.getOrGenerateDocumentID(entity)
	g.trySetIdentity(entity, id)
	return id
}

// Tries to set the identity property
func (g *GenerateEntityIDOnTheClient) trySetIdentity(entity interface{}, id string) {
	TrySetIDOnEntity(entity, id)
}
