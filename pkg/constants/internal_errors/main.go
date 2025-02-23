package internalerrors

import "errors"

var ErrNotFound = errors.New("not found")
var ErrBadRequest = errors.New("bad request")
var ErrUnauthorized = errors.New("unauthorized")
