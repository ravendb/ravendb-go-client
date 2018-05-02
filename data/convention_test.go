package data

import (
	"testing"
)

func TestWriteBehaviourSuccess(t *testing.T) {
	wb, err := NewWriteBehaviour(LEADER_ONLY)
	if wb.getBehaviourName() != "LeaderOnly" || err != nil {
		t.Fail()
	}

}

func TestWriteBehaviourFail(t *testing.T) {
	wb, err := NewWriteBehaviour(ROUND_ROBIN)
	if err == nil || wb != nil {
		t.Fail()
	}
}

func TestReadBehaviour(t *testing.T) {
	wb, err := NewReadBehaviour(FASTEST_NODE)
	if wb.getBehaviourName() != "FastestNode" || err != nil {
		t.Fail()
	}

}
