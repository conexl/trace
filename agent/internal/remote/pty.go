package remote

import "errors"

var ErrPTYNotImplemented = errors.New("remote PTY transport is reserved for the cloud/gRPC milestone")
