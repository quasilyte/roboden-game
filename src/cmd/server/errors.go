package main

import "errors"

var (
	errBadParams        = errors.New("bad params")
	errNotFound         = errors.New("data not found")
	errBadHTTPMethod    = errors.New("bad method")
	errQueueIsFull      = errors.New("queue is full")
	errUnsupportedBuild = errors.New("unsupported game build")
)

type archiveReason int

const (
	archiveUnknown archiveReason = iota
	archiveUnsupportedBuild
	archiveMismatchingResults
	archiveInvalidSeason
	archiveExecError
)
