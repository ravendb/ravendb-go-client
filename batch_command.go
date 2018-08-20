package ravendb

import (
	"bytes"
	"encoding/json"
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

type BatchCommand struct {
	RavenCommandBase

	_conventions       *DocumentConventions
	_commands          []ICommandData
	_attachmentStreams []io.Reader
	_options           *BatchOptions

	Result *JSONArrayResult
}

func NewBatchCommand(conventions *DocumentConventions, commands []ICommandData) (*BatchCommand, error) {
	return NewBatchCommandWithOptions(conventions, commands, nil)
}

func NewBatchCommandWithOptions(conventions *DocumentConventions, commands []ICommandData, options *BatchOptions) (*BatchCommand, error) {
	panicIf(conventions == nil, "conventions cannot be nil")
	panicIf(len(commands) == 0, "commands cannot be empty")

	cmd := &BatchCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_commands:    commands,
		_options:     options,
		_conventions: conventions,
	}

	for i := 0; i < len(commands); i++ {
		command := commands[i]
		if putAttachmentCommandData, ok := command.(*PutAttachmentCommandData); ok {

			stream := putAttachmentCommandData.getStream()
			for _, existingStream := range cmd._attachmentStreams {
				if stream == existingStream {
					return nil, throwStreamAlready()
				}
			}
			cmd._attachmentStreams = append(cmd._attachmentStreams, stream)
		}
	}

	return cmd, nil
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func (c *BatchCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/bulk_docs"
	url = c.appendOptions(url)

	var a []interface{}
	for _, cmd := range c._commands {
		el, err := cmd.serialize(c._conventions)
		if err != nil {
			return nil, err
		}
		a = append(a, el)
	}
	v := map[string]interface{}{
		"Commands": a,
	}
	js, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(c._attachmentStreams) == 0 {
		return NewHttpPost(url, js)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("main", string(js))

	nameCounter := 1
	for _, stream := range c._attachmentStreams {
		name := "attachment" + strconv.Itoa(nameCounter)
		nameCounter++
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"`, escapeQuotes(name)))
		h.Set("Command-Type", "AttachmentStream")
		part, err := writer.CreatePart(h)
		if err != nil {
			return nil, err
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
	req, err := NewHttpPostReader(url, body)
	if err != nil {
		return nil, err
	}
	contentType := writer.FormDataContentType()
	req.Header.Set("Content-Type", contentType)
	return req, nil
}

func (c *BatchCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return NewIllegalStateException("Got null response from the server after doing a batch, something is very wrong. Probably a garbled response.")
	}

	return json.Unmarshal(response, &c.Result)
}

func (c *BatchCommand) appendOptions(sb string) string {
	_options := c._options
	if _options == nil {
		return sb
	}

	sb += "?"

	if _options.isWaitForReplicas() {
		ts := TimeUtils_durationToTimeSpan(_options.getWaitForReplicasTimeout())
		sb += "&waitForReplicasTimeout=" + ts

		if _options.isThrowOnTimeoutInWaitForReplicas() {
			sb += "&throwOnTimeoutInWaitForReplicas=true"
		}

		sb += "&numberOfReplicasToWaitFor="
		if _options.isMajority() {
			sb += "majority"
		} else {
			sb += strconv.Itoa(_options.getNumberOfReplicasToWaitFor())
		}
	}

	if _options.isWaitForIndexes() {
		ts := TimeUtils_durationToTimeSpan(_options.getWaitForIndexesTimeout())
		sb += "&waitForIndexesTimeout=" + ts

		if _options.isThrowOnTimeoutInWaitForIndexes() {
			sb += "&waitForIndexThrow=true"
		} else {
			sb += "&waitForIndexThrow=false"
		}

		for _, specificIndex := range _options.getWaitForSpecificIndexes() {
			sb += "&waitForSpecificIndex=" + specificIndex
		}
	}
	return sb
}

func (c *BatchCommand) Close() {
	// empty
}

// Note: in Java is in PutAttachmentCommandHelper.java
func throwStreamAlready() error {
	return NewIllegalStateException("It is forbidden to re-use the same InputStream for more than one attachment. Use a unique InputStream per put attachment command.")
}
