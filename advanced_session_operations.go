package ravendb

import (
	"io"
	"reflect"
	"time"
)

// Note: Java's IAdvancedSessionOperations and IAdvancedDocumentSessionOperations
// is AdvancedSessionOperations

// AdvancedSessionOperations exposes advanced session operations
type AdvancedSessionOperations struct {
	s *DocumentSession
}

func (o *AdvancedSessionOperations) GetDocumentStore() *DocumentStore {
	return o.s.GetDocumentStore()
}

func (o *AdvancedSessionOperations) Attachments() *AttachmentsSessionOperations {
	return o.s.Attachments()
}

func (o *AdvancedSessionOperations) Revisions() *RevisionsSessionOperations {
	return o.s.Revisions()
}

func (o *AdvancedSessionOperations) Eagerly() *EagerSessionOperations {
	return o.s.Eagerly()
}

func (o *AdvancedSessionOperations) Lazily() *LazySessionOperations {
	return o.s.Lazily()
}

func (o *AdvancedSessionOperations) GetChangeVectorFor(instance interface{}) (*string, error) {
	return o.s.GetChangeVectorFor(instance)
}

func (o *AdvancedSessionOperations) GetMetadataFor(instance interface{}) (*MetadataAsDictionary, error) {
	return o.s.GetMetadataFor(instance)
}

func (o *AdvancedSessionOperations) GetRequestExecutor() *RequestExecutor {
	return o.s.GetRequestExecutor()
}

// GetNumberOfRequests returns number of requests sent to the server
func (o *AdvancedSessionOperations) GetNumberOfRequests() int {
	return o.s.GetNumberOfRequests()
}

func (o *AdvancedSessionOperations) Defer(commands ...ICommandData) {
	o.s.Defer(commands...)
}

func (o *AdvancedSessionOperations) Patch(entity interface{}, path string, value interface{}) error {
	return o.s.Patch(entity, path, value)
}

func (o *AdvancedSessionOperations) PatchByID(id string, path string, value interface{}) error {
	return o.s.PatchByID(id, path, value)
}

func (o *AdvancedSessionOperations) PatchArray(entity interface{}, pathToArray string, arrayAdder func(*JavaScriptArray)) error {
	return o.s.PatchArray(entity, pathToArray, arrayAdder)
}

func (o *AdvancedSessionOperations) PatchArrayByID(id string, pathToArray string, arrayAdder func(*JavaScriptArray)) error {
	return o.s.PatchArrayByID(id, pathToArray, arrayAdder)
}

func (o *AdvancedSessionOperations) Increment(entity interface{}, path string, valueToAdd interface{}) error {
	return o.s.Increment(entity, path, valueToAdd)
}

func (o *AdvancedSessionOperations) IncrementByID(id string, path string, valueToAdd interface{}) error {
	return o.s.IncrementByID(id, path, valueToAdd)
}

func (o *AdvancedSessionOperations) Refresh(entity interface{}) error {
	return o.s.Refresh(entity)
}

func (o *AdvancedSessionOperations) RawQuery(rawQuery string) *RawDocumentQuery {
	return o.s.RawQuery(rawQuery)
}

func (o *AdvancedSessionOperations) Query(opts *DocumentQueryOptions) *DocumentQuery {
	return o.s.Query(opts)
}

func (o *AdvancedSessionOperations) QueryCollection(collectionName string) *DocumentQuery {
	return o.s.QueryCollection(collectionName)
}

func (o *AdvancedSessionOperations) QueryCollectionForType(typ reflect.Type) *DocumentQuery {
	return o.s.QueryCollectionForType(typ)
}

func (o *AdvancedSessionOperations) QueryIndex(indexName string) *DocumentQuery {
	return o.s.QueryIndex(indexName)
}

func (o *AdvancedSessionOperations) StreamQuery(query *DocumentQuery, streamQueryStats *StreamQueryStatistics) (*StreamIterator, error) {
	return o.s.StreamQuery(query, streamQueryStats)
}

func (o *AdvancedSessionOperations) StreamRawQuery(query *RawDocumentQuery, streamQueryStats *StreamQueryStatistics) (*StreamIterator, error) {
	return o.s.StreamRawQuery(query, streamQueryStats)
}

func (o *AdvancedSessionOperations) StreamRawQueryInto(query *RawDocumentQuery, output io.Writer) error {
	return o.s.StreamRawQueryInto(query, output)
}

func (o *AdvancedSessionOperations) StreamQueryInto(query *DocumentQuery, output io.Writer) error {
	return o.s.StreamQueryInto(query, output)
}

func (o *AdvancedSessionOperations) Exists(id string) (bool, error) {
	return o.s.Exists(id)
}

func (o *AdvancedSessionOperations) WhatChanged() (map[string][]*DocumentsChanges, error) {
	return o.s.WhatChanged()
}

func (o *AdvancedSessionOperations) Evict(entity interface{}) error {
	return o.s.Evict(entity)
}

func (o *AdvancedSessionOperations) GetDocumentID(instance interface{}) string {
	return o.s.GetDocumentID(instance)
}

func (o *AdvancedSessionOperations) GetLastModifiedFor(instance interface{}) (*time.Time, error) {
	return o.s.GetLastModifiedFor(instance)
}

func (o *AdvancedSessionOperations) HasChanges() bool {
	return o.s.HasChanges()
}

func (o *AdvancedSessionOperations) HasChanged(entity interface{}) (bool, error) {
	return o.s.HasChanged(entity)
}

func (o *AdvancedSessionOperations) WaitForReplicationAfterSaveChanges(options func(*ReplicationWaitOptsBuilder)) {
	o.s.WaitForReplicationAfterSaveChanges(options)
}

func (o *AdvancedSessionOperations) WaitForIndexesAfterSaveChanges(options func(*IndexesWaitOptsBuilder)) {
	o.s.WaitForIndexesAfterSaveChanges(options)
}

func (o *AdvancedSessionOperations) IsLoaded(id string) bool {
	return o.s.IsLoaded(id)
}

func (o *AdvancedSessionOperations) IgnoreChangesFor(entity interface{}) {
	o.s.IgnoreChangesFor(entity)
}

func (o *AdvancedSessionOperations) Stream(args *StartsWithArgs) (*StreamIterator, error) {
	return o.s.Stream(args)
}

func (o *AdvancedSessionOperations) Clear() {
	o.s.Clear()
}

func (o *AdvancedSessionOperations) GetCurrentSessionNode() (*ServerNode, error) {
	return o.s.GetCurrentSessionNode()
}

func (o *AdvancedSessionOperations) AddBeforeStoreListener(handler func(*BeforeStoreEventArgs)) int {
	return o.s.AddBeforeStoreListener(handler)
}

func (o *AdvancedSessionOperations) RemoveBeforeStoreListener(handlerID int) {
	o.s.RemoveBeforeStoreListener(handlerID)
}

func (o *AdvancedSessionOperations) AddAfterSaveChangesListener(handler func(*AfterSaveChangesEventArgs)) int {
	return o.s.AddAfterSaveChangesListener(handler)
}

func (o *AdvancedSessionOperations) RemoveAfterSaveChangesListener(handlerID int) {
	o.s.RemoveAfterSaveChangesListener(handlerID)
}

func (o *AdvancedSessionOperations) AddBeforeDeleteListener(handler func(*BeforeDeleteEventArgs)) int {
	return o.s.AddBeforeDeleteListener(handler)
}

func (o *AdvancedSessionOperations) RemoveBeforeDeleteListener(handlerID int) {
	o.s.RemoveBeforeDeleteListener(handlerID)
}

func (o *AdvancedSessionOperations) AddBeforeQueryListener(handler func(*BeforeQueryEventArgs)) int {
	return o.s.AddBeforeQueryListener(handler)
}

func (o *AdvancedSessionOperations) RemoveBeforeQueryListener(handlerID int) {
	o.s.RemoveBeforeQueryListener(handlerID)
}

func (o *AdvancedSessionOperations) LoadStartingWith(results interface{}, args *StartsWithArgs) error {
	return o.s.LoadStartingWith(results, args)
}

func (o *AdvancedSessionOperations) LoadStartingWithIntoStream(output io.Writer, args *StartsWithArgs) error {
	return o.s.LoadStartingWithIntoStream(output, args)
}

func (o *AdvancedSessionOperations) LoadIntoStream(ids []string, output io.Writer) error {
	return o.s.LoadIntoStream(ids, output)
}

/*
int getMaxNumberOfRequestsPerSession();
void setMaxNumberOfRequestsPerSession(int maxRequests);
String storeIdentifier();
boolean isUseOptimisticConcurrency();
void setUseOptimisticConcurrency(boolean useOptimisticConcurrency);

EntityToJson getEntityToJson();
*/

/*
Map<String, Object> getExternalState();
*/
