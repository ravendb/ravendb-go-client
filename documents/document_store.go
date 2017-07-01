package documents

import "errors"

type DocumentStore struct{
	Urls []string
	DefaultDBName string
	initialized bool
}

func newDocumentStore(DefaultDBName string) (*DocumentStore, error){

	return &DocumentStore{}, nil
}

func (store DocumentStore) OpenSession() DocumentSession{
	return DocumentSession{}
}

func (store *DocumentStore) Initialize() error{
	if store.initialized {
		return nil
	}
	if err := store.validateConfiguration(); err != nil{
		return err
	}

}

func (store *DocumentStore) validateConfiguration() error{
	if store.Urls == nil{
		return errors.New("Store Urls is empty")
	}
	return nil
}


