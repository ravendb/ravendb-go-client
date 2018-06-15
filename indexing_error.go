package ravendb

import "time"

type IndexingError struct {
	Error     string    `json:"Error"`
	Timestamp time.Time `json:"Timestamp"`
	Document  string    `json:"Document"`
	Action    string    `json:"Action"`
}

func (e *IndexingError) getError() string {
	return e.Error
}

func (e *IndexingError) getTimestamp() time.Time {
	return e.Timestamp
}
func (e *IndexingError) getDocument() string {
	return e.Document
}

func (e *IndexingError) getAction() string {
	return e.Action
}

func (e *IndexingError) String() string {
	return "Error: " + e.Error + ", Document: " + e.Document + ", Action: " + e.Action
}

/*
public void setError(String error) {
	this.error = error;
}

public void setTimestamp(Date timestamp) {
	this.timestamp = timestamp;
}

public void setDocument(String document) {
	this.document = document;
}

public void setAction(String action) {
	this.action = action;
}
*/
