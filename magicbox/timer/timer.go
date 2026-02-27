package timer

import (
	"time"
)

type Timer interface {
	Match(time.Time) bool
}
