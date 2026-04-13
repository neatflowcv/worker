package domain

import "errors"

type BacklogItemStatus string

const (
	BacklogItemStatusOpen    BacklogItemStatus = "open"
	BacklogItemStatusRunning BacklogItemStatus = "running"
	BacklogItemStatusBlocked BacklogItemStatus = "blocked"
	BacklogItemStatusDone    BacklogItemStatus = "done"
)

var ErrInvalidBacklogItemStatus = errors.New("invalid backlog item status")

func (s BacklogItemStatus) validate() error {
	switch s {
	case BacklogItemStatusOpen,
		BacklogItemStatusRunning,
		BacklogItemStatusBlocked,
		BacklogItemStatusDone:
		return nil
	default:
		return ErrInvalidBacklogItemStatus
	}
}
