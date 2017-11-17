package raven_operations

import (
	"time"
	"github.com/ravendb-go-client/http/commands"
	SrvNodes "github.com/ravendb-go-client/http/server_nodes"
	"errors"
	"strconv"
	"fmt"
	"net/http"
	"github.com/ravendb-go-client/tools"
	"encoding/json"
	"github.com/ravendb-go-client/data"
)

//@param allow_stale Indicates whether operations are allowed on stale indexes.
//:type bool
//@param stale_timeout: If AllowStale is set to false and index is stale, then this is the maximum timeout to wait
//for index to become non-stale. If timeout is exceeded then exception is thrown.
//None by default - throw immediately if index is stale.
//:type timedelta
//@param max_ops_per_sec Limits the amount of base Operation per second allowed.
//:type int
//@param retrieve_details Determines whether Operation details about each document should be returned by server.
//:type bool
type QueryOperationOptions struct {
	allow_stale bool
	stale_timeout time.Duration
	max_ops_per_sec int
	retrieve_details bool
}
func NewQueryOperationOptions(allow_stale bool, stale_timeout time.Duration, max_ops_per_sec int, retrieve_details bool) *QueryOperationOptions {
	ref := &QueryOperationOptions{}
	ref.allow_stale = allow_stale
	ref.stale_timeout = stale_timeout
	ref.retrieve_details = retrieve_details
	ref.max_ops_per_sec = max_ops_per_sec
	
	return ref
}

type Operation struct {
	commands.RavenCommand
	operation string
}
func (ref *Operation) init() {
	ref.operation = "Operation"
}
func (ref *Operation) GetOperation() string{
	return ref.operation
}
type DeleteAttachmentOperation struct {
	Operation
	document_id, name, change_vector string
}
func NewDeleteAttachmentOperation(document_id, name, change_vector string) (*DeleteAttachmentOperation, error) {
	if document_id == "" {
		return nil, errors.New("Invalid documentId")
	}
	if name == "" {
		return nil, errors.New("Invalid name")
	}
	ref := &DeleteAttachmentOperation{}
	ref.init()
	ref.Method="DELETE"
	ref.document_id = document_id
	ref.name = name
	ref.change_vector = change_vector
	
	return ref, nil
}
func (ref *DeleteAttachmentOperation) create_request(sn SrvNodes.IServerNode) {
	ref.Url = fmt.Sprintf("%s/databases/%s/attachments?id=%s&name=%s", sn.GetUrl(), sn.GetDatabase(), strconv.Quote(ref.document_id),
		strconv.Quote(ref.name))
	if ref.change_vector > "" {
		ref.SetHeaders( map[string] string { "If-Match": `"` + ref.change_vector + `"` })
	}
}
//@param query_to_update: query that will be performed
//:type IndexQuery or str
//@param options: various Operation options e.g. AllowStale or MaxOpsPerSec
//:type QueryOperationOptions
//@return: json
//:rtype: dict of operation_id
type PatchByQueryOperation struct {
	Operation
	query_to_update *IndexQuery
	options *QueryOperationOptions
}
func NewPatchByQueryOperation(query_to_update *IndexQuery, options *QueryOperationOptions) (*PatchByQueryOperation, error) {
	if query_to_update == nil {
		return nil, errors.New("Invalid query")
	}
	if options == nil {
		options = NewQueryOperationOptions(true, 0, 0, false)
	}
	ref := &PatchByQueryOperation{options: options, query_to_update: query_to_update}
	ref.init()
	ref.Method="PATCH"

	return ref, nil
}
func (ref *PatchByQueryOperation) create_request(sn SrvNodes.IServerNode) {

	maxOpsPerSec := ref.options.max_ops_per_sec
	if maxOpsPerSec == 0 {
		maxOpsPerSec = ref.options.max_ops_per_sec
	}
	ref.Url = fmt.Sprintf("%s/databases/%s/queries?allowStale=%s&maxOpsPerSec=%d&details=%s", sn.GetUrl(), sn.GetDatabase(),
		ref.options.allow_stale, maxOpsPerSec, ref.options.retrieve_details)
	if ref.options.stale_timeout > 0 {
		ref.Url += "&staleTimeout=" + ref.options.stale_timeout.String()
	}

	ref.Data = map[string] interface{} { "Query": ref.query_to_update.to_json(), }
}
func (ref *Operation) GetResponseRaw(resp *http.Response) (out []byte, err error) {
	if resp == nil {
		return nil, errors.New("Invalid Response")
	}
//	response = response.json()
//	if "Error" in
//response:
//	return nil, errors.New(response["Error"])
//	return
//	{
//		"operation_id": response["OperationId"]
//	}
//	except
//ValueError:
//	raise
//	response.raise_for_status()
return
}

//@param query_to_delete: query that will be performed
//:type IndexQuery or str
//@param options: various Operation options e.g. AllowStale or MaxOpsPerSec
//:type QueryOperationOptions
//:rtype: dict of operation_id
type DeleteByQueryOperation struct {
	Operation
	query_to_delete *IndexQuery
	options *QueryOperationOptions
}
func NewDeleteByQueryOperation(query_to_delete *IndexQuery, options *QueryOperationOptions) (*DeleteByQueryOperation, error) {
	if query_to_delete == nil {
		return nil, errors.New("Invalid query")
	}

	if options == nil {
		options = NewQueryOperationOptions(true, 0, 0, false)
	}
	ref := &DeleteByQueryOperation{query_to_delete: query_to_delete, options: options}
	ref.init()
	ref.Method="DELETE"

	return ref, nil
}
func (ref *DeleteByQueryOperation) create_request(sn SrvNodes.IServerNode) {
	maxOpsPerSec := ref.options.max_ops_per_sec
	if maxOpsPerSec == 0 {
		maxOpsPerSec = ref.options.max_ops_per_sec
	}
	ref.Url = fmt.Sprintf("%s/databases/%s/queries?allowStale=%s&maxOpsPerSec=%d&details=%s", sn.GetUrl(), sn.GetDatabase(),
		ref.options.allow_stale, maxOpsPerSec, ref.options.retrieve_details)
	if ref.options.stale_timeout > 0 {
		ref.Url += "&staleTimeout=" + ref.options.stale_timeout.String()
	}
	ref.Data = map[string]interface{}{"Query": ref.query_to_delete.to_json(),}
}

func (ref *DeleteByQueryOperation) set_response(resp *http.Response) (out []byte, err error) {
	if resp == nil {
		// todo: where from the ref.index_name appeared???
		return nil, errors.New("Could not find index {ref.index_name}")
	}

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		return nil, errors.New(`tools.ResponseToJSON()["Error"]`)
	}
	return tools.ResponseToJSON(resp)
	//{
	//	"operation_id": response.json()["OperationId"]
	//}
}
type AttachmentType string
func (obj AttachmentType) String() string {
	return string(obj)
}
const (
document AttachmentType = "1"
revision = "2"
)

//@param documentId: The id of the document
//:type str
//@param name: The name of the attachment
//:type str
//@param attachment_type: The type of the attachment
//:type AttachmentType
//@param changeVector: The change vector of the document (needed only in revision)
//:type str
//@return: dict with the response and the attachment details

type GetAttachmentOperation struct {
	Operation
	document_id, name string
	attachment_type AttachmentType
	change_vector string
}
func NewGetAttachmentOperation(document_id, name string, attachment_type AttachmentType, change_vector string) (*GetAttachmentOperation, error) {
	if document_id == "" {
		return nil, errors.New("Invalid documentId")
	}
	if name == "" {
		return nil, errors.New("Invalid name")
	}

	if attachment_type != document && change_vector == ""{
		return nil, errors.New("Change Vector cannot be null for attachment type " + attachment_type.String())
	}

	ref := &GetAttachmentOperation{}
	ref.init()
	ref.Method = "GET"
	ref.document_id = document_id
	ref.name = name
	ref.attachment_type = attachment_type
	ref.change_vector = change_vector

	//super(GetAttachmentOperation._GetAttachmentCommand, ref..__init__(method="GET", is_read_request=True,
	//use_stream=True)
	return ref, nil
}
func (ref *GetAttachmentOperation) CreateRequest(sn SrvNodes.IServerNode) {
	ref.Url = fmt.Sprintf("%s/databases/%s/attachments?id=%s&name=%s", sn.GetUrl(), sn.GetDatabase(), strconv.Quote(ref.document_id),
		strconv.Quote(ref.name))

	if ref.attachment_type != document {
		ref.Method = "POST"
		ref.Data = map[string]string{"Type": string(ref.attachment_type), "ChangeVector": ref.change_vector,}
	}
}
type tAttachmentDetail struct {
	ContentType  []string `json:"contentType"`
	ChangeVector string `json:"changeVector"`
	Hash string `json:"hash"`
	Size string `json:"size"`
}
type tResponse struct {
	Response *http.Response `json:"response"`
	Details  tAttachmentDetail			`json:"details"`
}
func (ref *GetAttachmentOperation) GetResponseRaw(resp *http.Response) (out []byte, err error) {
	if resp == nil {
		return nil, nil
	}

	if resp.StatusCode == 200 {
		attachment_details := tAttachmentDetail{
			ContentType:  resp.Header["Content-Type"],
			ChangeVector: tools.GetChangeVectorFromHeader(resp),
			Hash:         resp.Header["Attachment-Hash"][0],
			Size:         resp.Header["Attachment-Size"][0],
		}

		return json.Marshal(tResponse{ Response: resp, Details: attachment_details})

	}
	return
}
//@param str documentId: The id of the document
//@param str changeVector: The changeVector
//@param PatchRequest patch: The patch that going to be applied on the document
//@param PatchRequest patchIfMissing: The default patch to applied
//@param bool skipPatchIfChangeVectorMismatch: If True will skip documents that mismatch the changeVector
type PatchOperation struct {
	Operation
	documentId, changeVector                                      string
	patch, patchIfMissing                                         data.PatchRequest
	skipPatchIfChangeVectorMismatch, returnDebugInformation, test bool
}
func NewPatchOperation(document_id, change_vector string, patch, patch_if_missing data.PatchRequest, skip_patch_if_change_vector_mismatch bool) *PatchOperation {
	ref := &PatchOperation{}
	ref.documentId = document_id
	ref.changeVector = change_vector
	ref.patch = patch
	ref.patchIfMissing = patch_if_missing
	ref.skipPatchIfChangeVectorMismatch = skip_patch_if_change_vector_mismatch
	return ref
}
//@param documentId: The id of the document
//@param name: Name of the attachment
//@param stream: The attachment as bytes (ex.open("file_path", "rb"))
//@param contentType: The type of the attachment (ex.image/png)
//@param changeVector: The change vector of the document
type PutAttachmentOperation struct {
	Operation
	stream                                                        []byte
	documentId, name, contentType, changeVector                   string
	skipPatchIfChangeVectorMismatch, returnDebugInformation, test bool
}
func NewPutAttachmentOperation(document_id, name string, stream []byte, content_type, change_vector string) *PutAttachmentOperation {
	ref := &PutAttachmentOperation{}
	ref.documentId = document_id
	ref.name = name
	ref.stream = stream
	ref.contentType = content_type
	ref.changeVector = change_vector

	return ref
}
//todo: implement
type FacetQuery struct {

}
//@param FacetQuery query: The query we wish to get
type GetFacetsOperation struct {
	Operation
	query *FacetQuery
}
func NewGetFacetsOperation(query *FacetQuery) (*GetFacetsOperation, error) {
	if query == nil {
		return nil, errors.New("Invalid query")
	}
	ref := &GetFacetsOperation{}
	ref.query = query
	//if query.wait_for_non_stale_results_timeout and query.wait_for_non_stale_results_timeout != timedelta.max:
	//self.timeout = self._query.wait_for_non_stale_results_timeout + timedelta(seconds=10)
	return ref, nil
}
type GetMultiFacetsOperation struct {
	Operation
	queries []FacetQuery
}
func NewGetMultiFacetsOperation(queries []FacetQuery) (*GetMultiFacetsOperation, error) {
	if queries == nil || len(queries) == 0 {
		return nil, errors.New("Invalid queries")
	}
	ref := &GetMultiFacetsOperation{}

	ref.queries = queries

	//super(GetMultiFacetsOperation._GetMultiFacetsCommand, ref..__init__(is_read_request=True)
	//requests = {}
	//for q in queries:
	//if not q:
	//return nil, errors.New("Invalid query")
	//requests.update(
	//{"url": "/queries", "query": "?op=facets&query-hash=" + q.get_query_hash(), "method": "POST",
	//"data": q.to_json()})
	return ref, nil
}
func (ref *GetMultiFacetsOperation) CreateRequest(serverNode SrvNodes.IServerNode) {
	fmt.Sprintf("%sref.command.CreateRequest(serverNode SrvNodes.IServerNode")
}
