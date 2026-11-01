package input

import "io"

type CreateGarmentWorkerInput struct {
	Name   string
	Phone  string
	Gender string
	Photo  *UploadFile
}

type UploadFile struct {
	Reader      io.Reader
	Size        int64
	ContentType string
	Filename    string
}
