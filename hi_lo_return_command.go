package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &HiLoReturnCommand{}
)

// HiLoReturnCommand represents "hi lo return" command
type HiLoReturnCommand struct {
	RavenCommandBase

	tag  string
	last int64
	end  int64
}

// NewHiLoReturnCommand returns a new HiLoReturnCommand
func NewHiLoReturnCommand(tag string, last int64, end int64) (*HiLoReturnCommand, error) {
	if last < 0 {
		return nil, newIllegalArgumentError("last is < 0")
	}
	if end < 0 {
		return nil, newIllegalArgumentError("end is < 0")
	}
	if tag == "" {
		return nil, newIllegalArgumentError("tag cannot be empty")
	}

	cmd := &HiLoReturnCommand{
		RavenCommandBase: NewRavenCommandBase(),

		tag:  tag,
		last: last,
		end:  end,
	}
	cmd.IsReadRequest = true
	cmd.ResponseType = RavenCommandResponseTypeEmpty
	return cmd, nil
}

func (c *HiLoReturnCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/hilo/return?tag=" + c.tag + "&end=" + i64toa(c.end) + "&last=" + i64toa(c.last)

	return NewHttpPut(url, nil)
}
