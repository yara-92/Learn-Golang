package model

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrInvalidCredential = errors.New("invalid username or password")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden: not the task owner")
	ErrTaskNotPending    = errors.New("task is not in pending status")
	ErrTemplateInvalid   = errors.New("template definition is invalid")
	ErrDuplicateUser     = errors.New("username already exists")
)
