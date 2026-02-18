package multiplexer

// OnExists constants define how to handle existing sessions.
const (
	OnExistsAttach = "attach"
	OnExistsKill   = "kill"
	OnExistsNew    = "new"
	OnExistsSwitch = "switch"
)

// ResolvedWindow represents a window/tab configuration after environment variable expansion.
type ResolvedWindow struct {
	Name    string
	Command string
	Cwd     string
}
