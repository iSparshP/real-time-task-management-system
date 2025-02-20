package task

import "errors"

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrInvalidStatus      = errors.New("invalid status")
	ErrInvalidPriority    = errors.New("invalid priority")
	ErrInvalidDueDate     = errors.New("invalid due date")
	ErrUnauthorized       = errors.New("unauthorized to perform this action")
	ErrDescriptionTooLong = errors.New("description exceeds maximum length")
	ErrInvalidAssignment  = errors.New("invalid task assignment")
	ErrInvalidPageSize    = errors.New("invalid page size")
	ErrInvalidSortField   = errors.New("invalid sort field")
	ErrInvalidTimeFormat  = errors.New("invalid time format")
)
