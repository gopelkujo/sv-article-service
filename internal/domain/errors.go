package domain

import "errors"

// ErrNotFound indicates the requested article does not exist.
var ErrNotFound = errors.New("article not found")
