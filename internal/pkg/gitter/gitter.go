package gitter

import "context"

type Gitter interface {
	CloneRepository(ctx context.Context, repositoryURL string, destinationDir string) error
	PullRepository(ctx context.Context, repositoryDir string) error
}
