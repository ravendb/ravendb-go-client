package ravendb

import (
	"fmt"
	"github.com/google/uuid"
)

type IRaftCommand interface {
	RaftUniqueRequestId() string
}

type RaftCommandBase struct {
	RavenCommandBase
	raftUniqueRequestId string
}

func (cmd *RaftCommandBase) RaftUniqueRequestId() string {
	if cmd.raftUniqueRequestId == "" {
		cmd.raftUniqueRequestId = fmt.Sprint(uuid.NewUUID())
	}
	return cmd.raftUniqueRequestId
}
