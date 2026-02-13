package node

import "encoding/json"

// NodeStatus определяет статус узла в кластере.
type NodeStatus int

const (
	NodeStatusConnecting NodeStatus = iota
	NodeStatusActive
	NodeStatusRemoved
)

func (status NodeStatus) String() string {
	switch status {
	case NodeStatusConnecting:
		return "Connecting"
	case NodeStatusActive:
		return "Active"
	case NodeStatusRemoved:
		return "Removed"
	default:
		return "Unknown"
	}
}

// MarshalJSON сериализует NodeStatus в JSON строку.
func (status NodeStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(status.String())
}

// UnmarshalJSON десериализует NodeStatus из JSON строки.
func (status *NodeStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "Connecting":
		*status = NodeStatusConnecting
	case "Active":
		*status = NodeStatusActive
	case "Removed":
		*status = NodeStatusRemoved
	default:
		*status = NodeStatusRemoved
	}

	return nil
}
