package diff

import (
	"errors"
	"time"
)

// ErrFullFetchRequired is returned when the cache cannot be updated from an incremental response
// and the next request should be a full fetch, i.e. without updatedAfter.
var ErrFullFetchRequired = errors.New("full fetch required")

// IncrementalDefinitionsReceiver is implemented by objects that support incremental definitions, i.e. responses
// where definitions are omitted (null) when unchanged.
// The client sets whether the control plane confirmed incremental definitions for the response and, if so,
// the cursor to be used as the next updatedAfter.
type IncrementalDefinitionsReceiver interface {
	SetIncrementalDefinitions(confirmed bool, cursor time.Time)
	IncrementalDefinitions() (confirmed bool, cursor time.Time)
}

// IncrementalDefinitionsState implements [IncrementalDefinitionsReceiver] and is meant to be embedded.
type IncrementalDefinitionsState struct {
	confirmed bool
	cursor    time.Time
}

func (s *IncrementalDefinitionsState) SetIncrementalDefinitions(confirmed bool, cursor time.Time) {
	s.confirmed, s.cursor = confirmed, cursor
}

func (s *IncrementalDefinitionsState) IncrementalDefinitions() (bool, time.Time) {
	return s.confirmed, s.cursor
}

// OmittableList is implemented by non updateable lists that can be omitted from incremental responses.
type OmittableList interface {
	// Omitted reports whether the list was omitted from the response, meaning it is unchanged.
	Omitted() bool
}
