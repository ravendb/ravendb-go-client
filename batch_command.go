package ravendb

import (
	"encoding/json"
	"io"
	"net/http"
)

type _BatchCommand struct {
	_conventions       *DocumentConventions
	_commands          []ICommandData
	_attachmentStreams []io.Reader
	_options           *BatchOptions
}

func NewBatchCommand(conventions *DocumentConventions, commands []ICommandData) *RavenCommand {
	return NewBatchCommandWithOptions(conventions, commands, nil)
}

func NewBatchCommandWithOptions(conventions *DocumentConventions, commands []ICommandData, options *BatchOptions) *RavenCommand {
	data := &_BatchCommand{
		_commands:    commands,
		_options:     options,
		_conventions: conventions,
	}
	panicIf(conventions == nil, "conventions cannot be nil")
	panicIf(len(commands) == 0, "commands cannot be empty")

	for i := 0; i < len(commands); i++ {
		command := commands[i]
		if putAttachmentCommandData, ok := command.(*PutAttachmentCommandData); ok {

			stream := putAttachmentCommandData.getStream()
			// TODO: verify no duplicate stream
			//			if !_attachmentStreams.add(stream) {
			//	PutAttachmentCommandHelper.throwStreamAlready()
			//}
			data._attachmentStreams = append(data._attachmentStreams, stream)
		}
	}

	cmd := NewRavenCommand()
	cmd.data = data
	cmd.createRequestFunc = BatchCommand_createRequest
	cmd.setResponseFunc = BatchCommand_setResponse
	return cmd
}

func BatchCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, string) {
	data := cmd.data.(*_BatchCommand)
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/bulk_docs"
	// TODO: appendOptions(sb)
	var a []interface{}
	for _, cmd := range data._commands {
		el, err := cmd.serialize(data._conventions)
		must(err) // TODO: return
		a = append(a, el)
	}
	v := map[string]interface{}{
		"Commands": a,
	}
	js, err := json.Marshal(v)
	must(err)
	return NewHttpPost(url, string(js)), url
}

func BatchCommand_setResponse(cmd *RavenCommand, response String, fromCache bool) error {
	if response == "" {
		return NewIllegalStateException("Got null response from the server after doing a batch, something is very wrong. Probably a garbled response.")
	}

	var res JSONArrayResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	cmd.result = &res
	return nil
}
