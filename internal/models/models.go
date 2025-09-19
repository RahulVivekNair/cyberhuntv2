package models

import (
	"time"
)

type Group struct {
	ID             int
	Name           string
	Pathway        string
	CurrentClueIdx int
	Completed      bool
	EndTime        *time.Time
	Password       string
}

type Clue struct {
	ID      int
	Pathway string
	Index   int
	Content string
	QRCode  string
}

type GameSettings struct {
	ID          int
	TotalClues  int
	StartTime   *time.Time
	GameStarted bool
	GameEnded   bool
}

type Admin struct {
	ID int
}
