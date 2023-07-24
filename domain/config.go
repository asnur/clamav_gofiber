package domain

type Config struct {
	ClamdAddress string // ClamdAddress is the address of the clamd daemon
	FieldName    string // FieldName is the name of the field in the multipart form
}
