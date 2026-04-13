package domain

type Project struct {
	id            string
	name          string
	repositoryURL string
	auth          *Auth
}

func NewProject(id, name, repositoryURL string, auth *Auth) *Project {
	return &Project{
		id:            id,
		name:          name,
		repositoryURL: repositoryURL,
		auth:          auth,
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

func (p *Project) Auth() *Auth {
	return p.auth
}
