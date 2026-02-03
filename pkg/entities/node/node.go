package node

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
