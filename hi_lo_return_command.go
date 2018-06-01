package ravendb

import (
	"net/http"
	"strconv"
)

type _HiLoReturnCommand struct {
	_tag  String
	_last int
	_end  int
}

func NewHiLoReturnCommand(tag String, last int, end int) *RavenCommand {
	panicIf(last < 0, "last is < 0")
	panicIf(end < 0, "end is < 0")
	panicIf(tag == "", "tag cannot be empty")

	data := &_HiLoReturnCommand{
		_tag:  tag,
		_last: last,
		_end:  end,
	}

	cmd := NewRavenCommand()
	cmd.IsReadRequest = true
	cmd.data = data
	cmd.createRequestFunc = HiLoReturnCommand_createRequest

	return cmd
}

func HiLoReturnCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, string) {
	data := cmd.data.(*_HiLoReturnCommand)
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/hilo/return?tag=" + data._tag + "&end=" + strconv.Itoa(data._end) + "&last=" + strconv.Itoa(data._last)

	return NewHttpPut(url, ""), url
}
