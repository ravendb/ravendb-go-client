package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
)

const (
	proxyURL = "http://localhost:8888"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func runSingleJavaTest(className string, logFileTmpl string) {
	go proxy.Run(logFileTmpl)
	defer proxy.CloseLogFile()

	// Running just one maven test: https://stackoverflow.com/a/18136440/2898
	// mvn -Dtest=HiLoTest test
	cmd := exec.Command("mvn", fmt.Sprintf("-Dtest=%s", className), "test")
	cmd.Dir = path.Join("..", "ravendb-jvm-client")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	must(err)
	err = cmd.Wait()
	must(err)
}

func runJava() {
	// TODO: for some reason when we run more than one in a sequence,
	// the second one fails. Possibly because the server fails
	// to start the second time

	//runSingleJavaTest("HiLoTest", "trace_hilo_java.txt")
	//runSingleJavaTest("GetTopologyTest", "trace_get_topology_java.txt")
	//runSingleJavaTest("GetTcpInfoTest", "trace_get_tcp_info_java.txt")
	//runSingleJavaTest("GetClusterTopologyTest", "trace_get_cluster_topology_java.txt")
	//runSingleJavaTest("DeleteTest", "trace_delete_java.txt")
	//runSingleJavaTest("RequestExecutorTest", "trace_request_executor_java.txt")
	//runSingleJavaTest("ExistsTest", "trace_exists_java.txt")
	//runSingleJavaTest("ClientConfigurationTest", "trace_client_configuration_java.txt")
	//runSingleJavaTest("LoadTest", "trace_load_java.txt")
	//runSingleJavaTest("TrackEntityTest", "trace_track_entity_java.txt")
	//runSingleJavaTest("PutDocumentCommandTest", "trace_put_document_command_java.txt")
	//runSingleJavaTest("GetStatisticsCommandTest", "trace_get_statistics_java.txt")
	//runSingleJavaTest("DeleteDocumentCommandTest", "trace_delete_document_command_java.txt")
	//runSingleJavaTest("GetNextOperationIdCommandTest", "trace_get_next_operation_id_java.txt")
	//runSingleJavaTest("CompactTest", "trace_compact_java.txt")
	//runSingleJavaTest("NextAndSeedIdentitiesTest", "trace_next_and_seed_identities_java.txt")
	//runSingleJavaTest("StoreTest", "trace_store_java.txt")
	//runSingleJavaTest("ContainsTest", "trace_contains_java.txt")
	//runSingleJavaTest("RevisionsTest", "trace_revisions_java.txt")
	//runSingleJavaTest("PatchTest", "trace_patch_java.txt")
	//runSingleJavaTest("IndexOperationsTest", "trace_index_operations_java.txt")
	//runSingleJavaTest("documents.LoadTest", "trace_documents_load_java.txt")
	//runSingleJavaTest("DeleteByQueryTest", "trace_delete_by_query_java.txt")
	//runSingleJavaTest("CrudTest", "trace_crud_java.txt")
	//runSingleJavaTest("UniqueValuesTest", "trace_unique_values_java.txt")
	//runSingleJavaTest("AttachmentsRevisionsTest", "trace_attachments_revisions_java.txt")
	//runSingleJavaTest("BasicDocumentsTest", "trace_basic_documents_java.txt")
	//runSingleJavaTest("WhatChangedTest", "trace_what_changed_java.txt")
	//runSingleJavaTest("BulkInsertsTest", "trace_bulk_inserts_java.txt")
	//runSingleJavaTest("AttachmentsSessionTest", "trace_attachments_session_java.txt")
	//runSingleJavaTest("QueryTest", "trace_query_java.txt")
	//runSingleJavaTest("IndexesFromClientTest", "trace_indexes_from_client_java.txt")
	runSingleJavaTest("RegexQueryTest", "trace_regex_query_java.txt")
}

func main() {
	runJava()
}
