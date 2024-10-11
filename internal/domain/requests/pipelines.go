package requests

import "time"

type PipelineStatus string

const (
	PipelineFailed  = PipelineStatus("failed")
	PipelineSuccess = PipelineStatus("success")
)

type Pipeline struct {
	ID         int
	IID        int
	ProjectID  int
	Status     PipelineStatus
	UpdatedAt  *time.Time
	CreatedAt  *time.Time
	StartedAt  *time.Time
	FinishedAt *time.Time
}
