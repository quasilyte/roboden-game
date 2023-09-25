package staging

import "errors"

var (
	errIllegalAction      = errors.New("illegal action")
	errInvalidColonyIndex = errors.New("invalid colony index")
	errExcessiveAcions    = errors.New("excessive actions")
	errBadCheckpoint      = errors.New("mismatching checkpoint value")
)
