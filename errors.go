package main

import "errors"

var (
	ErrExit         = errors.New("exit qry")
	ErrReadLineFail = errors.New("failed to initialize prompt")
)
