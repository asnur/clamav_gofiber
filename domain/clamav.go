package domain

const (
	RES_OK          = "OK"
	RES_FOUND       = "FOUND"
	RES_ERROR       = "ERROR"
	RES_PARSE_ERROR = "PARSE ERROR"
)

type ScanResult struct {
	Raw         string
	Description string
	Path        string
	Hash        string
	Size        int
	Status      string
}
