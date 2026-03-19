package apiserver

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

var (
	ErrNotFound          = connect.NewError(connect.CodeNotFound, fmt.Errorf("entity not found"))
	ErrFieldMaskRequired = connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("field mask is required"))
)

func newConnectError(err error) *connect.Error {
	if errors.Is(err, storecmn.ErrNotFound) {
		return ErrNotFound
	}
	return connect.NewError(connect.CodeInternal, err)
}

func newConnectErrorWithCode(code connect.Code, err error) *connect.Error {
	if errors.Is(err, storecmn.ErrNotFound) {
		return ErrNotFound
	}
	return connect.NewError(code, err)
}
