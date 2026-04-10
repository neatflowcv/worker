package domain

type BacklogItemStatus string

const (
	BacklogItemStatusOpen    BacklogItemStatus = "open"
	BacklogItemStatusRunning BacklogItemStatus = "running"
	BacklogItemStatusBlocked BacklogItemStatus = "blocked"
	BacklogItemStatusDone    BacklogItemStatus = "done"
)

type BacklogItem struct {
	id          string
	projectID   string
	title       string
	description string
	status      BacklogItemStatus
	// See repository docs: backlog-ordering and ADR 0001.
	orderKey string
}

func NewBacklogItem(id, projectID, title, description, orderKey string) *BacklogItem {
	return &BacklogItem{
		id:          id,
		projectID:   projectID,
		title:       title,
		description: description,
		status:      BacklogItemStatusOpen,
		orderKey:    orderKey,
	}
}

func (i *BacklogItem) ID() string {
	return i.id
}

func (i *BacklogItem) ProjectID() string {
	return i.projectID
}

func (i *BacklogItem) Title() string {
	return i.title
}

func (i *BacklogItem) Description() string {
	return i.description
}

func (i *BacklogItem) Status() BacklogItemStatus {
	return i.status
}

func (i *BacklogItem) OrderKey() string {
	return i.orderKey
}
