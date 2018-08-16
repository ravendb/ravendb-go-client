package ravendb

type IndexErrors struct {
	Name   string           `json:"Name"`
	Errors []*IndexingError `json:"Errors"`
}

func NewIndexErrors() *IndexErrors {
	return &IndexErrors{}
}

func (e *IndexErrors) GetName() string {
	return e.Name
}

func (e *IndexErrors) getErrors() []*IndexingError {
	return e.Errors
}

/*
public void setName(String name) {
	this.name = name;
}

public void setErrors(IndexingError[] errors) {
	this.errors = errors;
}
*/
