package models

type CommandType int

const (
	PutCommand CommandType = iota
	DeleteCommand
)

type Command struct {
	Type  CommandType
	Key   string
	Value string
}

type StateMachine interface {
	Apply(command Command)
}

type LogEntry struct {
	Term    int32
	Command string
}
