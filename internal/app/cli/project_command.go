package cli

type projectCommand struct {
	Backlog projectBacklogCommand `cmd:"" help:"Manage project backlogs." name:"backlog"`
	Create  projectCreateCommand  `cmd:"" help:"Create a project."        name:"create"`
	List    projectListCommand    `cmd:"" help:"List projects."           name:"list"`
}
