package services

import "errors"

var ErrGroupExists = errors.New("this group already exists")
var ErrGameAlreadyStarted = errors.New("game already started")
var ErrNoSettingsRow = errors.New("no game settings row found")
var ErrGameAlreadyEnded = errors.New("game has already ended")
var ErrGameNotStarted = errors.New("game has not started yet")
