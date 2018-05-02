package server_nodes

import "time"

type IServerNode interface {
	GetClusterTag() string
	SetResponseTime(duration time.Duration)
	GetUrl() string
	GetDatabase() string
}

//Its better to make fields private esp considering that there are setters there
type ServerNode struct {
	Url, Database string
	ClusterTag    string
	ResponseTime  time.Duration
}

func NewServerNode(url string, database string) (*ServerNode, error) {
	return &ServerNode{url, database, "", 0}, nil
}

func (node ServerNode) GetClusterTag() string {
	return node.ClusterTag
}

func (node ServerNode) SetResponseTime(duration time.Duration) {
	node.ResponseTime = duration
}

func (node ServerNode) GetUrl() string {
	return node.Url
}

func (node ServerNode) GetDatabase() string {
	return node.Database
}
