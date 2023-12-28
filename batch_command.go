package ravendb

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
)

var (
	_ RavenCommand = &BatchCommand{}
)

// BatchCommand represents batch command
type BatchCommand struct {
	RavenCommandBase

	conventions       *DocumentConventions
	commands          []ICommandData
	attachmentStreams []io.Reader
	options           *BatchOptions
	Result            *JSONArrayResult

	transactionMode             int
	disableAtomicDocumentWrites *bool
	raftUniqueRequestId         string
}

// newBatchCommand returns new BatchCommand
func newBatchCommand(conventions *DocumentConventions, commands []ICommandData, options *BatchOptions, transactionMode int, disableAtomicDocumentWrites *bool) (*BatchCommand, error) {
	if conventions == nil {
		return nil, newIllegalStateError("conventions cannot be nil")
	}
	if len(commands) == 0 {
		return nil, newIllegalStateError("commands cannot be empty")
	}

	raftId, err := "", error(nil)

	if transactionMode == TransactionMode_ClusterWide {
		raftId, err = RaftId()

		if err != nil {
			return nil, err
		}
	}

	cmd := &BatchCommand{
		RavenCommandBase: NewRavenCommandBase(),

		commands:    commands,
		options:     options,
		conventions: conventions,

		transactionMode:             transactionMode,
		disableAtomicDocumentWrites: disableAtomicDocumentWrites,
		raftUniqueRequestId:         raftId,
	}

	for i := 0; i < len(commands); i++ {
		command := commands[i]
		if putAttachmentCommandData, ok := command.(*PutAttachmentCommandData); ok {

			stream := putAttachmentCommandData.getStream()
			for _, existingStream := range cmd.attachmentStreams {
				if stream == existingStream {
					return nil, throwStreamAlready()
				}
			}
			cmd.attachmentStreams = append(cmd.attachmentStreams, stream)
		}

	}

	return cmd, nil
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func (c *BatchCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/bulk_docs"
	url = c.appendOptions(url)

	var a []interface{}
	for _, cmd := range c.commands {
		el, err := cmd.serialize(c.conventions)
		if err != nil {
			return nil, err
		}
		a = append(a, el)
	}

	v := map[string]interface{}{
		"Commands": a,
	}

	if c.transactionMode == TransactionMode_ClusterWide {
		v["TransactionMode"] = "ClusterWide"
	}

	js, err := jsonMarshal(v)
	if err != nil {
		return nil, err
	}
	if len(c.attachmentStreams) == 0 {
		return NewHttpPost(url, js)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	err = writer.WriteField("main", string(js))
	if err != nil {
		return nil, err
	}

	nameCounter := 1
	for _, stream := range c.attachmentStreams {
		name := "attachment" + strconv.Itoa(nameCounter)
		nameCounter++
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"`, escapeQuotes(name)))
		h.Set("Command-Type", "AttachmentStream")
		// Note: Java seems to set those by default
		h.Set("Content-Type", "application/octet-stream")
		h.Set("Content-Transfer-Encoding", "binary")

		part, err2 := writer.CreatePart(h)
		if err2 != nil {
			return nil, err2
		}
		_, err = io.Copy(part, stream)
		if err != nil {
			return nil, err
		}
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	req, err := newHttpPostReader(url, body)
	if err != nil {
		return nil, err
	}
	contentType := writer.FormDataContentType()
	req.Header.Set("Content-Type", contentType)

	return req, nil
}

func (c *BatchCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return newIllegalStateError("Got null response from the server after doing a batch, something is very wrong. Probably a garbled response.")
	}

	return jsonUnmarshal(response, &c.Result)
}

func (c *BatchCommand) appendOptions(sb string) string {
	_options := c.options
	if _options == nil && c.transactionMode == TransactionMode_SingleNode {
		return sb
	}

	sb += "?"

	if c.transactionMode == TransactionMode_ClusterWide {
		if c.disableAtomicDocumentWrites != nil {
			sb += "&disableAtomicDocumentWrites="
			if *c.disableAtomicDocumentWrites == false {
				sb += "false"
			} else {
				sb += "true"
			}
		}

		sb += "&raft-request-id=" + c.raftUniqueRequestId

		if _options == nil {
			return sb
		}
	}

	if _options.waitForReplicas {
		ts := durationToTimeSpan(_options.waitForReplicasTimeout)
		sb += "&waitForReplicasTimeout=" + ts

		if _options.throwOnTimeoutInWaitForReplicas {
			sb += "&throwOnTimeoutInWaitForReplicas=true"
		}

		sb += "&numberOfReplicasToWaitFor="
		if _options.majority {
			sb += "majority"
		} else {
			sb += strconv.Itoa(_options.numberOfReplicasToWaitFor)
		}
	}

	if _options.waitForIndexes {
		ts := durationToTimeSpan(_options.waitForIndexesTimeout)
		sb += "&waitForIndexesTimeout=" + ts

		if _options.throwOnTimeoutInWaitForIndexes {
			sb += "&waitForIndexThrow=true"
		} else {
			sb += "&waitForIndexThrow=false"
		}

		for _, specificIndex := range _options.waitForSpecificIndexes {
			sb += "&waitForSpecificIndex=" + specificIndex
		}
	}

	return sb
}

func (c *BatchCommand) Close() error {
	// no-op
	return nil
}

// Note: in Java is in PutAttachmentCommandHelper.java
func throwStreamAlready() error {
	return newIllegalStateError("It is forbidden to re-use the same InputStream for more than one attachment. Use a unique InputStream per put attachment command.")
}
