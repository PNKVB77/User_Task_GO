package models

import "errors"

var ErrorNotFound = errors.New("task not found")
var ErrorNoRows = errors.New("no such rows")
var ErrorInvalidInput = errors.New("invalid input")
