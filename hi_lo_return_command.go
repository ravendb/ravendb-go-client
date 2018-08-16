package ravendb

import (
	"net/http"
	"strconv"
)

var (
	_ RavenCommand = &HiLoReturnCommand{}
)

type HiLoReturnCommand struct {
	*RavenCommandBase

	_tag  string
	_last int
	_end  int
}

func NewHiLoReturnCommand(tag string, last int, end int) *HiLoReturnCommand {
	panicIf(last < 0, "last is < 0")
	panicIf(end < 0, "end is < 0")
	panicIf(tag == "", "tag cannot be empty")

	cmd := &HiLoReturnCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_tag:  tag,
		_last: last,
		_end:  end,
	}
	cmd.IsReadRequest = true
	cmd.ResponseType = RavenCommandResponseType_EMPTY
	return cmd
}

func (c *HiLoReturnCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/hilo/return?tag=" + c._tag + "&end=" + strconv.Itoa(c._end) + "&last=" + strconv.Itoa(c._last)

	return NewHttpPut(url, nil)
}
