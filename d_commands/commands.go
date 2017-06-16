package d_commands



type RavenCommand struct{
	url string
	method string
	data string
	headers string
	isRavenCommand bool
	isReadRequest bool
	failedNodes []int
	authRetries int
	avoidFailover bool
}

type GetDocumentCommand struct{

	ravenCommand RavenCommand

	//The key of the documents you want to retrieve
	keys []string

	//Array of paths in documents in which server should look for a 'referenced' document
	includes []string

	//Specifies if only document metadata should be returned
	metadataOnly bool
}

type DeleteDocumentCommand struct{

	ravenCommand RavenCommand

	//unique key under which document will be deleted
	key string

	//current document etag, used for concurrency checks (nil to skip check)
	etag int64
}

type PutDocumentCommand struct{

	ravenCommand RavenCommand

	//unique key under which document will be stored
	key string

	//document data
	document []interface{}

	//current document etag, used for concurrency checks (nil to skip check)
	etag int64
}