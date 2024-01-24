package types

type AttackerCommand int

const (
	CMD_NULL AttackerCommand = iota
	CMD_CONTINUE
	CMD_RETURN
	CMD_ABORT
	CMD_EXIT
)

type AttackerResponse struct {
	Cmd    AttackerCommand `json:"cmd"`
	Result string          `json:"result"`
}
