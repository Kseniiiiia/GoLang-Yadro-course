package core

import "context"

type Normalizer interface {
	Norm(context.Context, string) ([]string, error)
}

type Pinger interface {
	Ping(context.Context) error
}

type Updater interface {
	Update(context.Context) error
	Stats(context.Context) (UpdateStats, error)
	Status(context.Context) (UpdateStatus, error)
	Drop(context.Context) error
}

type Searcher interface {
	Search(context.Context, string, int32) ([]Comics, int32, error)
	IndexSearch(context.Context, string, int32) ([]Comics, int32, error)
}

type YoloDetector interface {
	Detect(ctx context.Context, imageData []byte) ([]Yolo, error)
}
