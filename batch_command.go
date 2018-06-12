package ravendb

import (
	"encoding/json"
	"io"
	"net/http"
)

var (
	_ RavenCommand = &BatchCommand{}
)

type BatchCommand struct {
	*RavenCommandBase

	_conventions       *DocumentConventions
	_commands          []ICommandData
	_attachmentStreams []io.Reader
	_options           *BatchOptions

	Result *JSONArrayResult
}

func NewBatchCommand(conventions *DocumentConventions, commands []ICommandData) *BatchCommand {
	return NewBatchCommandWithOptions(conventions, commands, nil)
}

func NewBatchCommandWithOptions(conventions *DocumentConventions, commands []ICommandData, options *BatchOptions) *BatchCommand {
	panicIf(conventions == nil, "conventions cannot be nil")
	panicIf(len(commands) == 0, "commands cannot be empty")

	cmd := &BatchCommand{
		RavenCommandBase: NewRavenCommandBase(),
		_commands:        commands,
		_options:         options,
		_conventions:     conventions,
	}

	for i := 0; i < len(commands); i++ {
		command := commands[i]
		if putAttachmentCommandData, ok := command.(*PutAttachmentCommandData); ok {

			stream := putAttachmentCommandData.getStream()
			// TODO: verify no duplicate stream
			//			if !_attachmentStreams.add(stream) {
			//	PutAttachmentCommandHelper.throwStreamAlready()
			//}
			cmd._attachmentStreams = append(cmd._attachmentStreams, stream)
		}
	}

	return cmd
}

func (c *BatchCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/bulk_docs"
	// TODO: appendOptions(sb)
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
	return NewHttpPost(url, string(js))
}

func (c *BatchCommand) setResponse(response String, fromCache bool) error {
	if response == "" {
		return NewIllegalStateException("Got null response from the server after doing a batch, something is very wrong. Probably a garbled response.")
	}

	var res JSONArrayResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
