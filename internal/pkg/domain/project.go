package domain

type Project struct {
	id            string
	name          string
	repositoryURL string
}

func NewRepository(id, name, repositoryURL string) *Project {
	return &Project{
		id:            id,
		name:          name,
		repositoryURL: repositoryURL,
	}
}

func (p *Project) ID() string {
	return p.id
}

func (p *Project) Name() string {
	return p.name
}

func (p *Project) RepositoryURL() string {
	return p.repositoryURL
}
