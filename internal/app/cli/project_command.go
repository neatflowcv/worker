package cli

type projectCommand struct {
	Create projectCreateCommand `cmd:"" help:"Create a project." name:"create"`
	List   projectListCommand   `cmd:"" help:"List projects."    name:"list"`
}
