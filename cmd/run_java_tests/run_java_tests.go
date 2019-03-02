package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/kjk/httplogproxy"
)

const (
	proxyURL = "http://localhost:8888"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func isUpper(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

// converts "IndexesFromClientTest" => "indexes_from_client"
func testNameToFileName(s string) string {
	s = strings.TrimSuffix(s, "Test")
	lower := strings.ToLower(s)
	var res []byte
	n := len(s)
	for i := 0; i < n; i++ {
		c := s[i]
		if i > 0 && isUpper(c) {
			res = append(res, '_')
		}
		res = append(res, lower[i])
	}
	return string(res)
}

func runSingleJavaTest(className string) {
	logFileName := "trace_" + testNameToFileName(className) + "_java.txt"
	logFilePath := filepath.Join("logs", logFileName)
	go httplogproxy.Run(logFilePath)
	defer httplogproxy.CloseLogFile()

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

	//runSingleJavaTest("AdvancedPatchingTest")
	//runSingleJavaTest("AggregationTest")
	//runSingleJavaTest("AttachmentsSessionTest")
	//runSingleJavaTest("AttachmentsRevisionsTest")
	//runSingleJavaTest("BasicDocumentsTest")
	//runSingleJavaTest("BulkInsertsTest")
	//runSingleJavaTest("ClientConfigurationTest")
	//runSingleJavaTest("CompactTest")
	//runSingleJavaTest("ContainsTest")
	//runSingleJavaTest("CrudTest")
	//runSingleJavaTest("DeleteTest")
	//runSingleJavaTest("DeleteDocumentCommandTest")
	//runSingleJavaTest("DocumentsLoadTest")
	//runSingleJavaTest("DeleteByQueryTest")
	//runSingleJavaTest("ExistsTest")
	//runSingleJavaTest("GetTopologyTest")
	//runSingleJavaTest("GetTcpInfoTest")
	//runSingleJavaTest("GetClusterTopologyTest")
	//runSingleJavaTest("GetStatisticsCommandTest")
	//runSingleJavaTest("GetNextOperationIDCommandTest")
	//runSingleJavaTest("HiLoTest")
	//runSingleJavaTest("IndexOperationsTest")
	//runSingleJavaTest("IndexesFromClientTest")
	//runSingleJavaTest("LoadIntoStreamTest")
	//runSingleJavaTest("LoadTest")
	//runSingleJavaTest("NextAndSeedIdentitiesTest")
	//runSingleJavaTest("PatchTest")
	//runSingleJavaTest("PutDocumentCommandTest")
	//runSingleJavaTest("QueryTest")
	//runSingleJavaTest("RegexQueryTest")
	//runSingleJavaTest("RequestExecutorTest")
	//runSingleJavaTest("RevisionsTest")
	//runSingleJavaTest("StoreTest")
	//runSingleJavaTest("TrackEntityTest")
	//runSingleJavaTest("UniqueValuesTest")
	//runSingleJavaTest("SpatialSortingTest")
	//runSingleJavaTest("SpatialTest")
	//runSingleJavaTest("SpatialQueriesTest")
	//runSingleJavaTest("SpatialSearchTest")
	//runSingleJavaTest("WhatChangedTest")
	//runSingleJavaTest("WhatChangedTest")
	//runSingleJavaTest("FirstClassPatchTest")
	//runSingleJavaTest("RavenDB_8761")
	//runSingleJavaTest("RavenDB_10641Test")
	//runSingleJavaTest("SimonBartlettTest")
	//runSingleJavaTest("RavenDB_9676Test")
	//runSingleJavaTest("RavenDB_5669Test")
	//runSingleJavaTest("RavenDB903Test")
	//runSingleJavaTest("CustomSerializationTest")
	//runSingleJavaTest("QueriesWithCustomFunctionsTest")
	//runSingleJavaTest("SuggestionsTest")
	//runSingleJavaTest("MoreLikeThisTest")
	//runSingleJavaTest("ChangesTest")
	runSingleJavaTest("DocumentStreaming")
	//runSingleJavaTest("QueryStreaming")
	//runSingleJavaTest("LoadAllStartingWith")
	//runSingleJavaTest("SuggestionsLazyTest")
	//runSingleJavaTest("LazyTest")
	//runSingleJavaTest("LazyAggregationEmbedded")
	//runSingleJavaTest("AggressiveCaching")
	//runSingleJavaTest("CachingOfDocumentInclude")
	//runSingleJavaTest("RevisionsSubscriptionsTest")
	//runSingleJavaTest("SubscriptionsBasicTest")
	//runSingleJavaTest("SecuredSubscriptionsBasicTest")
	//runSingleJavaTest("FacetPagingTest")
}

func main() {
	runJava()
}
