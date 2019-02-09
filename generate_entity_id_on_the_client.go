package ravendb

import "strings"

type genIDFunc func(interface{}) (string, error)

type generateEntityIDOnTheClient struct {
	_conventions *DocumentConventions
	_generateID  genIDFunc
}

func newGenerateEntityIDOnTheClient(conventions *DocumentConventions, generateID genIDFunc) *generateEntityIDOnTheClient {
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
func (g *generateEntityIDOnTheClient) getOrGenerateDocumentID(entity interface{}) (string, error) {
	var err error
	id, ok := g.tryGetIDFromInstance(entity)
	if !ok || id == "" {
		id, err = g._generateID(entity)
		if err != nil {
			return "", err
		}
	}

	if strings.HasPrefix(id, "/") {
		return "", newIllegalStateError("Cannot use value '" + id + "' as a document id because it begins with a '/'")
	}
	return id, nil
}

func (g *generateEntityIDOnTheClient) generateDocumentKeyForStorage(entity interface{}) (string, error) {
	id, err := g.getOrGenerateDocumentID(entity)
	if err != nil {
		return "", err
	}
	g.trySetIdentity(entity, id)
	return id, nil
}

// Tries to set the identity property
func (g *generateEntityIDOnTheClient) trySetIdentity(entity interface{}, id string) {
	TrySetIDOnEntity(entity, id)
}
