package domain

import "errors"

type BacklogItem struct {
	id          string
	projectID   string
	title       string
	description string
	status      BacklogItemStatus
	// See repository docs: backlog-ordering and ADR 0001.
	orderKey string
}

var ErrBacklogItemTitleRequired = errors.New("backlog item title is required")
var ErrBacklogItemCannotStart = errors.New("cannot start backlog item")
var ErrBacklogItemCannotBlock = errors.New("cannot block backlog item")
var ErrBacklogItemCannotResume = errors.New("cannot resume backlog item")

func NewBacklogItem(
	id, projectID, title, description string,
	status BacklogItemStatus,
	orderKey string,
) (*BacklogItem, error) {
	if title == "" {
		return nil, ErrBacklogItemTitleRequired
	}

	err := status.validate()
	if err != nil {
		return nil, err
	}

	return &BacklogItem{
		id:          id,
		projectID:   projectID,
		title:       title,
		description: description,
		status:      status,
		orderKey:    orderKey,
	}, nil
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

func (i *BacklogItem) SetDescription(description string) *BacklogItem {
	item := i.clone()
	item.description = description

	return item
}

func (i *BacklogItem) SetTitle(title string) (*BacklogItem, error) {
	if title == "" {
		return nil, ErrBacklogItemTitleRequired
	}

	item := i.clone()
	item.title = title

	return item, nil
}

func (i *BacklogItem) Start() (*BacklogItem, error) {
	if i.status != BacklogItemStatusOpen {
		return nil, ErrBacklogItemCannotStart
	}

	item := i.clone()
	item.status = BacklogItemStatusRunning

	return item, nil
}

func (i *BacklogItem) Blocked() (*BacklogItem, error) {
	if i.status != BacklogItemStatusRunning {
		return nil, ErrBacklogItemCannotBlock
	}

	item := i.clone()
	item.status = BacklogItemStatusBlocked

	return item, nil
}

func (i *BacklogItem) Resume() (*BacklogItem, error) {
	if i.status != BacklogItemStatusBlocked {
		return nil, ErrBacklogItemCannotResume
	}

	item := i.clone()
	item.status = BacklogItemStatusRunning

	return item, nil
}

func (i *BacklogItem) clone() *BacklogItem {
	return &BacklogItem{
		id:          i.id,
		projectID:   i.projectID,
		title:       i.title,
		description: i.description,
		status:      i.status,
		orderKey:    i.orderKey,
	}
}
