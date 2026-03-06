package errorhandling

import (
	"errors"
)

var ErrNoRowsAffected = errors.New("no rows affected")
var ErrTooManyRequests = errors.New("too many requests")