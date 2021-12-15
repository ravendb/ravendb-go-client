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
	assert.NotNil(t, store)
	defer store.Close()

	path := os.Getenv("RAVENDB_TEST_CA_PATH")
	if !fileExists(path) {
		fmt.Printf("Didn't find cert.crt file at '%s'. Set RAVENDB_TEST_CA_PATH env variable\n", path)
		os.Exit(1)
	}

	certificate := loadTestCaCertificate(path)
	assert.NotNil(t, certificate)
	fmt.Printf("Loaded client certificate from '%s'\n", path)

	certName := "Admin Certificate"
	operation := certificates.OperationPutCertificate{
		CertName:          certName,
		CertBytes:         certificate.Raw,
		SecurityClearance: certificates.ClusterAdmin.String(),
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
