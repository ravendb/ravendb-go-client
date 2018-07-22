package ravendb

import "strings"

var _ QueryToken = &FromToken{}

type FromToken struct {
	collectionName string
	indexName      string
	dynamic        bool
	alias          string
}

func (t *FromToken) getCollectionName() string {
	return t.collectionName
}

func (t *FromToken) getIndexName() string {
	return t.indexName
}

func (t *FromToken) isDynamic() bool {
	return t.dynamic
}

func (t *FromToken) getAlias() string {
	return t.alias
}

func NewFromToken(indexName string, collectionName string, alias string) *FromToken {
	return &FromToken{
		collectionName: collectionName,
		indexName:      indexName,
		dynamic:        collectionName != "",
		alias:          alias,
	}
}

func FromToken_create(indexName string, collectionName string, alias string) *FromToken {
	return NewFromToken(indexName, collectionName, alias)
}

func (t *FromToken) writeTo(writer *StringBuilder) {
	if t.indexName == "" && t.collectionName == "" {
		panicIf(true, "Either indexName or collectionName must be specified")
		// NewIllegalStateException("Either indexName or collectionName must be specified");
	}

	if t.dynamic {
		writer.append("from ")

		hasWhitespace := (strings.IndexAny(t.collectionName, " \t\r\n") != -1)
		if hasWhitespace {
			if strings.Contains(t.collectionName, "\"") {
				t.throwInvalidCollectionName()
			}
			writer.append('"').append(t.collectionName).append('"')
		} else {
			QueryToken_writeField(writer, t.collectionName)
		}
	} else {
		writer.append("from index '")
		writer.append(t.indexName)
		writer.append("'")
	}

	if t.alias != "" {
		writer.append(" as ").append(t.alias)
	}
}

func (t *FromToken) throwInvalidCollectionName() {
	panicIf(true, "Collection name cannot contain a quote, but was: %s", t.collectionName)
	// NewIllegalArgumentException("Collection name cannot contain a quote, but was: " + collectionName);
}
