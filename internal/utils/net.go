package utils

import (
	"errors"
	"io"
	"net"
)

func IsErrTimeout(err error) bool {
	if errors.Is(err, io.EOF) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	return false
}
