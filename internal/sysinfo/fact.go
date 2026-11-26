package sysinfo

// Fact identifies a piece of data sysinfo can collect independently of
// the others. The identifiers are stable: configuration files refer to
// the items built from them.
type Fact string

// State says whether a fact was found.
type State uint8

// The states of a fact. Absent and Unknown are kept apart on purpose:
// a machine without a battery has nothing to show, while a battery whose
// files cannot be read is worth a row that says so.
const (
	// Unknown means the fact was not requested, its collection failed or
	// it timed out.
	Unknown State = iota
	// Present means the fact was collected.
	Present
	// Absent means the fact does not exist on this machine.
	Absent
)

// String returns the name of the state, for test output.
func (s State) String() string {
	switch s {
	case Present:
		return "present"
	case Absent:
		return "absent"
	default:
		return "unknown"
	}
}

// Value is a fact of type T together with its state. V is meaningful only
// when State is Present. The zero value is Unknown.
type Value[T any] struct {
	State State
	V     T
}

// present returns a Value that holds v.
func present[T any](v T) Value[T] {
	return Value[T]{State: Present, V: v}
}

// absent returns a Value that records that the fact does not exist.
func absent[T any]() Value[T] {
	return Value[T]{State: Absent}
}
