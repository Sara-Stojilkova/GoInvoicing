package apperrors

import "errors"

var ErrNotFound  = errors.New("not found")
var ErrConflict  = errors.New("already exists")
var ErrForbidden = errors.New("forbidden")
