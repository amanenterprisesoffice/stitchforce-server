package input

import (
	"io"
	"stitchforce-server/internal/entity"
)

type CreateGarmentWorkerInput struct {
	Name   string
	Phone  string
	Gender string
	Photo  *UploadFile
}

type UpdateGarmentWorkerInput struct {
	Name   string
	Phone  string
	Gender string
	Photo  *UploadFile
	Status entity.WorkerStatus
}

type UploadFile struct {
	Reader      io.Reader
	Size        int64
	ContentType string
	Filename    string
}
