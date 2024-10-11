package requests

import "time"

type Discussion struct {
	ID             string
	IndividualNote bool
	Notes          []*Note
}

type Note struct {
	ID         int
	Author     User
	CreatedAt  *time.Time
	Resolvable bool
	Resolved   bool
}
