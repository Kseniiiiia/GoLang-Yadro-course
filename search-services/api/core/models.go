package core

type UpdateStatus string

const (
	StatusUpdateUnknown UpdateStatus = "unknown"
	StatusUpdateIdle    UpdateStatus = "idle"
	StatusUpdateRunning UpdateStatus = "running"
)

type UpdateStats struct {
	WordsTotal    int
	WordsUnique   int
	ComicsFetched int
	ComicsTotal   int
}

type Comics struct {
	ID    int
	URL   string
	Score int
}

type Yolo struct {
	BBox       []float32 `json:"bbox"`
	Confidence float32   `json:"confidence"`
	Label      string    `json:"label"`
	LabelNum   int       `json:"label_num"`
}
