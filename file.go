package conveyearthgo

import (
	"time"
)

type File struct {
	ID      int64
	Message int64
	Hash    string
	Mime    string
	Created time.Time
}
