package transaction

import "trellis.tech/trellis/common.v1/errcode"

var (
	ErrAtLeastOneRepo   = errcode.New("input one repo at least")
	ErrNotFoundFunction = errcode.New("not found function")
	ErrFailToCreateRepo = errcode.New("fail to create an new repo")
	ErrNotFoundEngine   = errcode.New("not found engine")
)
