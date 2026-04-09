package cli

type app struct {
	Project projectCommand `cmd:"" help:"Manage projects." name:"project"`
}

func newApp() *app {
	var ret app

	return &ret
}
