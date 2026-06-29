package analysis

import "go/types"

// Env is the source environment of a reference.
type Env int

const (
	EnvProduction Env = iota
	EnvTest
)

// UsageState tracks in which environments a declaration was referenced.
type UsageState struct {
	Production bool
	Test       bool
}

// UsageIndex maps declarations to their usage state, keyed by types.Object.
// A position-based fallback handles cases where the same source file is
// compiled into both a regular package and a test-compiled variant; within a
// single packages.Load session each file has a stable token.Pos, so the same
// declaration has the same Pos regardless of which compilation produced the
// types.Object.
type UsageIndex struct {
	byObject map[types.Object]*UsageState
	byPos    map[int64]*UsageState // token.Pos (int64) -> state
}

func newUsageIndex() *UsageIndex {
	return &UsageIndex{
		byObject: make(map[types.Object]*UsageState),
		byPos:    make(map[int64]*UsageState),
	}
}

func (u *UsageIndex) Mark(obj types.Object, env Env) {
	s := u.stateFor(obj)
	switch env {
	case EnvProduction:
		s.Production = true
	case EnvTest:
		s.Test = true
	}
}

func (u *UsageIndex) Get(obj types.Object) UsageState {
	if s, ok := u.byObject[obj]; ok {
		return *s
	}
	if s, ok := u.byPos[int64(obj.Pos())]; ok {
		return *s
	}
	return UsageState{}
}

func (u *UsageIndex) stateFor(obj types.Object) *UsageState {
	if s, ok := u.byObject[obj]; ok {
		return s
	}
	// Check by position (handles test-compiled variants sharing source positions).
	pos := int64(obj.Pos())
	if s, ok := u.byPos[pos]; ok {
		u.byObject[obj] = s // cache for fast future lookups
		return s
	}
	s := &UsageState{}
	u.byObject[obj] = s
	u.byPos[pos] = s
	return s
}
