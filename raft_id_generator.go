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

func (cmd *RaftCommandBase) RaftUniqueRequestId() (string, error) {
	if cmd.raftUniqueRequestId == "" {
		newUUID, err := uuid.NewUUID()
		if err != nil {
			return "", err
		}
		cmd.raftUniqueRequestId = fmt.Sprintf("%v", newUUID)
	}
	return cmd.raftUniqueRequestId, nil
}
