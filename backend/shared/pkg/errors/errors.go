package errors

import (
	"connectrpc.com/connect"
)

func NewNotFoundError(msg string) error {
	return connect.NewError(connect.CodeNotFound, nil)
}

func NewInvalidArgumentError(msg string) error {
	return connect.NewError(connect.CodeInvalidArgument, nil)
}

func NewInternalError(msg string) error {
	return connect.NewError(connect.CodeInternal, nil)
}

func NewUnauthenticatedError(msg string) error {
	return connect.NewError(connect.CodeUnauthenticated, nil)
}
