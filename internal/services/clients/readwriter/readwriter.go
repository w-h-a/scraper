package readwriter

import (
	"github.com/w-h-a/scraper/internal/services/clients/reader"
	"github.com/w-h-a/scraper/internal/services/clients/writer"
)

type ReadWriterType string

const (
	Mock   ReadWriterType = "mock"
	Sheets ReadWriterType = "sheets"
)

var (
	ReadWriterTypes = map[string]ReadWriterType{
		"mock":   Mock,
		"sheets": Sheets,
	}
)

type ReadWriter interface {
	reader.Reader
	writer.Writer
}
