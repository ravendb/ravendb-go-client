package certificates

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/ravendb/ravendb-go-client"
	"net/http"
)

type OperationPutCertificate struct {
	CertName          string            `json:"CertName,omitempty"`
	CertBytes         []byte            `json:"CertBytes,omitempty"`
	SecurityClearance string            `json:"SecurityClearance,omitempty"`
	Permissions       map[string]string `json:"Permissions,omitempty"`
}

func (operation *OperationPutCertificate) GetCommand(conventions *ravendb.DocumentConventions) (ravendb.RavenCommand, error) {
	return &putCertificateCommand{
		RaftCommandBase: ravendb.RaftCommandBase{
			RavenCommandBase: ravendb.RavenCommandBase{
				ResponseType: ravendb.RavenCommandResponseTypeObject,
			},
		},
		parent: operation,
	}, nil
}

type putCertificateCommand struct {
	ravendb.RaftCommandBase
	parent *OperationPutCertificate
}

func (c *putCertificateCommand) CreateRequest(node *ravendb.ServerNode) (*http.Request, error) {
	raftUniqueRequestId, err := c.RaftCommandBase.RaftUniqueRequestId()
	if err != nil {
		return nil, err
	}
	url := node.URL + "/admin/certificates?raft-request-id=" + raftUniqueRequestId

	var body map[string]interface{}
	body = make(map[string]interface{})
	body["Name"] = c.parent.CertName
	body["Certificate"] = base64.StdEncoding.EncodeToString(c.parent.CertBytes)
	body["SecurityClearance"] = c.parent.SecurityClearance
	body["Permissions"] = c.parent.Permissions

	bodyMarshalled, err := json.MarshalIndent(body, "", "\t")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(bodyMarshalled))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	return req, nil

}

func (c *putCertificateCommand) SetResponse(response []byte, fromCache bool) error {
	return json.Unmarshal(response, c.parent)
}
