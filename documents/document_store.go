package documents

type DocumentStore struct{

}

func (store DocumentStore) OpenSession() DocumentSession{
	return DocumentSession{}
}


