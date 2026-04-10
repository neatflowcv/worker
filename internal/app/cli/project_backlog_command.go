package cli

type projectBacklogCommand struct {
	Abort     projectBacklogAbortCommand     `cmd:"" help:"Abort a backlog item."      name:"abort"`
	Breakdown projectBacklogBreakdownCommand `cmd:"" help:"Break down a backlog item." name:"breakdown"`
	Create    projectBacklogCreateCommand    `cmd:"" help:"Create a backlog item."     name:"create"`
	Done      projectBacklogDoneCommand      `cmd:"" help:"Complete a backlog item."   name:"done"`
	Feedback  projectBacklogFeedbackCommand  `cmd:"" help:"Send backlog feedback."     name:"feedback"`
	List      projectBacklogListCommand      `cmd:"" help:"List backlog items."        name:"list"`
	Get       projectBacklogGetCommand       `cmd:"" help:"Get a backlog item."        name:"get"`
	Refine    projectBacklogRefineCommand    `cmd:"" help:"Refine a backlog item."     name:"refine"`
	Start     projectBacklogStartCommand     `cmd:"" help:"Start a backlog item."      name:"start"`
	Update    projectBacklogUpdateCommand    `cmd:"" help:"Update a backlog item."     name:"update"`
	Move      projectBacklogMoveCommand      `cmd:"" help:"Move a backlog item."       name:"move"`
	Delete    projectBacklogDeleteCommand    `cmd:"" help:"Delete a backlog item."     name:"delete"`
}
