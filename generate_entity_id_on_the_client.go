package ravendb

type genIDFunc func(interface{}) string

type generateEntityIDOnTheClient struct {
	_conventions *DocumentConventions
	_generateID  genIDFunc
}

func newgenerateEntityIDOnTheClient(conventions *DocumentConventions, generateID genIDFunc) *generateEntityIDOnTheClient {
	return &generateEntityIDOnTheClient{
		_conventions: conventions,
		_generateID:  generateID,
	}
}

// Attempts to get the document key from an instance
func (g *generateEntityIDOnTheClient) tryGetIDFromInstance(entity interface{}) (string, bool) {
	panicIf(entity == nil, "Entity cannot be null")
	return tryGetIDFromInstance(entity)
}

// Tries to get the identity.
func (g *generateEntityIDOnTheClient) getOrGenerateDocumentID(entity interface{}) string {
	id, ok := g.tryGetIDFromInstance(entity)
	if !ok || id == "" {
		id = g._generateID(entity)
	}

	/* TODO:
	        if (id != null && id.startsWith("/")) {
	            throw new IllegalStateException("Cannot use value '" + id + "' as a document id because it begins with a '/'");
			}
	*/
	return id
}

func (g *generateEntityIDOnTheClient) generateDocumentKeyForStorage(entity interface{}) string {
	id := g.getOrGenerateDocumentID(entity)
	g.trySetIdentity(entity, id)
	return id
}

// Tries to set the identity property
func (g *generateEntityIDOnTheClient) trySetIdentity(entity interface{}, id string) {
	TrySetIDOnEntity(entity, id)
}
