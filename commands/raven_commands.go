package commands

import (
	"errors"
	"net/http"
	"time"
)

type RavenCommand struct {
	Url, Method, Data                      string
	Headers                                map[string]string
	isReadRequest, useStream, ravenCommand bool
	failedNodes                            []string
	timeout                                time.Duration
	requested_node                         string
}

func (ref *RavenCommand) Init() {
	if ref.Headers == nil {
		ref.Headers = make(map[string]string, 0)
	}
}
func (ref *RavenCommand) createRequest(serverNode string) {
	panic(errors.New("NotImplementedError"))
}

func (ref *RavenCommand) GetResponseRaw(resp *http.Response) {
	panic(errors.New("NotImplementedError"))
}

func (obj RavenCommand) RavenCommand() bool {
	return obj.ravenCommand

}

func (obj RavenCommand) IsFailedWithNode(node string) bool {
	for _, val := range obj.failedNodes {
		if val == node {
			return true
		}
	}
	return false
}
