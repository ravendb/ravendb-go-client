package ravendb

type GenIDFunc func(interface{}) string

type GenerateEntityIdOnTheClient struct {
	_conventions *DocumentConventions
	_generateId  GenIDFunc
}

func NewGenerateEntityIdOnTheClient(conventions *DocumentConventions, generateId GenIDFunc) *GenerateEntityIdOnTheClient {
	return &GenerateEntityIdOnTheClient{
		_conventions: conventions,
		_generateId:  generateId,
	}
}

// Attempts to get the document key from an instance
func (g *GenerateEntityIdOnTheClient) tryGetIdFromInstance(entity interface{}) (string, bool) {
	panicIf(entity == nil, "Entity cannot be null")
	return tryGetIdFromInstance(entity)
}

// Tries to get the identity.
func (g *GenerateEntityIdOnTheClient) getOrGenerateDocumentId(entity interface{}) string {
	id, ok := g.tryGetIdFromInstance(entity)
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

func (g *GenerateEntityIdOnTheClient) generateDocumentKeyForStorage(entity interface{}) string {
	id := g.getOrGenerateDocumentId(entity)
	g.trySetIdentity(entity, id)
	return id
}

// Tries to set the identity property
func (g *GenerateEntityIdOnTheClient) trySetIdentity(entity interface{}, id string) {
	TrySetIDOnEntity(entity, id)
}
