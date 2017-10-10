package commands

import(
	".."
	"fmt"
	gohttp "net/http"
	"encoding/json"
	"time"
)

type GetTopologyCommand struct{
	forcedUrl string
	ravenCommand http.RavenCommand
	Result http.Topology
}

func NewGetTopologyCommand(forcedUrl string) (*GetTopologyCommand, error){
	ravenCommand, err := http.NewRavenCommand()
	ravenCommand.FailedNodes = make(map[http.IServerNode]error)
	if err != nil{
		return nil, err
	}
	return &GetTopologyCommand{forcedUrl, *ravenCommand, nil}, nil
}

func (command GetTopologyCommand) CreateRequest(node http.IServerNode, urlPtr *string) (*gohttp.Request, error){
	*urlPtr = fmt.Sprintf("%s/topology?name=%s", node, *urlPtr)
	if command.forcedUrl != ""{
		*urlPtr += fmt.Sprintf("&url=%s", command.forcedUrl)
	}
	return gohttp.NewRequest("GET", *urlPtr, nil)
}

func (command GetTopologyCommand) SetResponse(response string, fromCache bool) error{
	if response == ""{
		return nil
	}
	if err := json.Unmarshal([]byte(response), command.ravenCommand.Result); err != nil {
		return err
	}
}

func (command GetTopologyCommand) GetTimeout() time.Duration{
	return command.ravenCommand.GetTimeout()
}

func (command GetTopologyCommand) SetFailedNode(node http.IServerNode, err error){
	command.ravenCommand.FailedNodes[node] = err
}

func (command GetTopologyCommand) Send(client gohttp.Client, request *gohttp.Request) (*gohttp.Response, error){
	return command.ravenCommand.Send(client, request)
}

func (command GetTopologyCommand) SetStatusCode(code int){
	command.ravenCommand.SetStatusCode(code)
}

func (command GetTopologyCommand) GetFailedNodes() map[http.IServerNode]error{
	return command.ravenCommand.GetFailedNodes()
}

func (command GetTopologyCommand) ShouldRefreshTopology() bool{
	return command.ravenCommand.ShouldRefreshTopology()
}

func (command GetTopologyCommand) ProcessResponse(response gohttp.Response, url string){
	command.ravenCommand.ProcessResponse(response, url)
}