package domain

type Auth struct {
	username string
	password string
}

func NewAuth(username, password string) *Auth {
	return &Auth{
		username: username,
		password: password,
	}
}

func (a *Auth) Username() string {
	return a.username
}

func (a *Auth) Password() string {
	return a.password
}
