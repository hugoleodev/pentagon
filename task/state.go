package task

import (
	"encoding/json"
	"fmt"
	"strings"
)

type State int

const (
	Pending State = iota
	Scheduled
	Running
	Completed
	Failed
)

var (
	State_names = map[State]string{
		Pending:   "pending",
		Scheduled: "scheduled",
		Running:   "running",
		Completed: "completed",
		Failed:    "failed",
	}

	State_values = map[string]State{
		"pending":   Pending,
		"scheduled": Scheduled,
		"running":   Running,
		"completed": Completed,
		"failed":    Failed,
	}
)

func (s State) String() string {
	return State_names[s]
}

func ParseState(s string) (State, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	value, ok := State_values[s]
	if !ok {
		return State(0), fmt.Errorf("%q is not a valid state", s)
	}
	return State(value), nil
}

func (s State) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *State) UnmarshalJSON(data []byte) error {
	var state string
	var err error

	if err = json.Unmarshal(data, &state); err != nil {
		return err
	}
	if *s, err = ParseState(state); err != nil {
		return err
	}

	return nil
}

var stateTransitionMap = map[State][]State{
	Pending:   {Scheduled},
	Scheduled: {Scheduled, Running, Failed},
	Running:   {Running, Completed, Failed},
	Completed: {},
	Failed:    {},
}

func Contains(states []State, state State) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}

func ValidStateTransition(src State, dst State) bool {
	return Contains(stateTransitionMap[src], dst)
}
