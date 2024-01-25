package main

import (
	"time"

	"github.com/mfonism/charmed/connections/internals/sets"
)

type WordGroup struct {
	members sets.Set[string]
	clue    string
	color   Color
	RevelationStatus
}

func newWordGroup(members sets.Set[string], clue string, color Color) WordGroup {
	return WordGroup{members: members, clue: clue, color: color}
}

type Color int

const (
	Yellow Color = iota + 1
	Green
	Blue
	Purple
)

type RevelationStatus struct {
	revealer Revealer
	unix     int64
}

type Revealer int

const (
	None Revealer = iota
	Player
	Computer
)

func (rs *RevelationStatus) isUnrevealed() bool {
	return rs.revealer == None
}

func (rs *RevelationStatus) isRevealedByPlayer() bool {
	return rs.revealer == Player
}

func (rs *RevelationStatus) makeRevealedByPlayer() {
	rs.revealer = Player
	rs.unix = time.Now().Unix()
}

func (rs *RevelationStatus) isRevealedByComputer() bool {
	return rs.revealer == Computer
}

func (rs *RevelationStatus) makeRevealedByComputer() {
	rs.revealer = Computer
	rs.unix = time.Now().Unix()
}
