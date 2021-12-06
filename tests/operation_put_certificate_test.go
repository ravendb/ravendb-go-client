package tests

import (
	"fmt"
	"github.com/ravendb/ravendb-go-client/serverwide/certificates"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func putCertificateTest(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getSecuredDocumentStoreMust(t)
	defer store.Close()

	path := os.Getenv("RAVENDB_TEST_CLIENT_CERTIFICATE_PATH")
	if !fileExists(path) {
		fmt.Printf("Didn't find cert.pem file at '%s'. Set RAVENDB_TEST_CLIENT_CERTIFICATE_PATH env variable\n", path)
		os.Exit(1)
	}
	certificate := loadTestCaCertificate(path)
	assert.NotNil(t, certificate)
	fmt.Printf("Loaded client certificate from '%s'\n", path)

	operation := certificates.OperationPutCertificate{
		CertName: "Admin Certificate",
		//CertBytes:         certificate.Raw,
		SecurityClearance: certificates.SecurityClearance.Operator,
		Permissions:       nil,
	}
	err = store.Maintenance().Server().Send(&operation)
	assert.NoError(t, err)

}

func TestPutCertificateTest(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() {
		destroyDriver(t, driver)
	}

	defer recoverTest(t, destroy)

	putCertificateTest(t, driver)
}
