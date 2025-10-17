package config

import (
	"os"
	"sync"

	"github.com/w-h-a/scraper/internal/clients/readwriter"
	"github.com/w-h-a/scraper/internal/clients/scraper"
)

var (
	instance *config
	once     sync.Once
)

type config struct {
	env                         string
	name                        string
	version                     string
	logsAddress                 string
	logsAPIKeyHeader            string
	logsAPIKeyValue             string
	tracesAddress               string
	tracesAPIKeyHeader          string
	tracesAPIKeyValue           string
	scraper                     string
	readwriter                  string
	readwriterLocation          string
	sheetsServiceAccountKeyPath string
}

func New() {
	once.Do(func() {
		instance = &config{
			env:                         "dev",
			name:                        "golang-job-scraper",
			version:                     "0.1.0-alpha.0",
			logsAddress:                 "api.honeycomb.io",
			logsAPIKeyHeader:            "x-honeycomb-team",
			logsAPIKeyValue:             "",
			tracesAddress:               "api.honeycomb.io",
			tracesAPIKeyHeader:          "x-honeycomb-team",
			tracesAPIKeyValue:           "",
			scraper:                     "feed",
			readwriter:                  "sheets",
			readwriterLocation:          "",
			sheetsServiceAccountKeyPath: "service_account_key.json",
		}

		env := os.Getenv("ENV")
		if len(env) > 0 {
			instance.env = env
		}

		name := os.Getenv("NAME")
		if len(name) > 0 {
			instance.name = name
		}

		version := os.Getenv("VERSION")
		if len(version) > 0 {
			instance.version = version
		}

		logsAddress := os.Getenv("LOGS_ADDRESS")
		if len(logsAddress) > 0 {
			instance.logsAddress = logsAddress
		}

		logsAPIKeyHeader := os.Getenv("LOGS_API_KEY_HEADER")
		if len(logsAPIKeyHeader) > 0 {
			instance.logsAPIKeyHeader = logsAPIKeyHeader
		}

		logsAPIKeyValue := os.Getenv("LOGS_API_KEY_VALUE")
		if len(logsAPIKeyValue) > 0 {
			instance.logsAPIKeyValue = logsAPIKeyValue
		}

		tracesAddress := os.Getenv("TRACES_ADDRESS")
		if len(tracesAddress) > 0 {
			instance.tracesAddress = tracesAddress
		}

		tracesAPIKeyHeader := os.Getenv("TRACES_API_KEY_HEADER")
		if len(tracesAPIKeyHeader) > 0 {
			instance.tracesAPIKeyHeader = tracesAPIKeyHeader
		}

		tracesAPIKeyValue := os.Getenv("TRACES_API_KEY_VALUE")
		if len(tracesAPIKeyValue) > 0 {
			instance.tracesAPIKeyValue = tracesAPIKeyValue
		}

		s := os.Getenv("SCRAPER")
		if len(s) > 0 {
			if _, ok := scraper.ScraperTypes[s]; ok {
				instance.scraper = s
			} else {
				panic("unsupported scraper")
			}
		}

		rw := os.Getenv("READ_WRITER")
		if len(rw) > 0 {
			if _, ok := readwriter.ReadWriterTypes[rw]; ok {
				instance.readwriter = rw
			} else {
				panic("unsupported readwriter")
			}
		}

		readwriterLocation := os.Getenv("READ_WRITER_LOCATION")
		if len(readwriterLocation) > 0 {
			instance.readwriterLocation = readwriterLocation
		}

		sheetsServiceAccountKeyPath := os.Getenv("SHEETS_SERVICE_ACCOUNT_KEY_PATH")
		if len(sheetsServiceAccountKeyPath) > 0 {
			instance.sheetsServiceAccountKeyPath = sheetsServiceAccountKeyPath
		}
	})
}

func Env() string {
	if instance == nil {
		panic("cfg is nil")
	}

	return instance.env
}

func Name() string {
	if instance == nil {
		panic("cfg is nil")
	}

	return instance.name
}

func Version() string {
	if instance == nil {
		panic("cfg is nil")
	}

	return instance.version
}

func LogsAddress() string {
	if instance == nil {
		panic("cfg is nil")
	}

	return instance.logsAddress
}

func LogsAPIKeyHeader() string {
	if instance == nil {
		panic("cfg is nil")
	}

	return instance.logsAPIKeyHeader
}

func LogsAPIKeyValue() string {
	if instance == nil {
		panic("cfg is nil")
	}

	return instance.logsAPIKeyValue
}

func TracesAddress() string {
	if instance == nil {
		panic("cfg is nil")
	}

	return instance.tracesAddress
}

func TracesAPIKeyHeader() string {
	if instance == nil {
		panic("cfg is nil")
	}

	return instance.tracesAPIKeyHeader
}

func TracesAPIKeyValue() string {
	if instance == nil {
		panic("cfg is nil")
	}

	return instance.tracesAPIKeyValue
}

func Scraper() string {
	if instance == nil {
		panic("cfg is nil")
	}

	return instance.scraper
}

func ReadWriter() string {
	if instance == nil {
		panic("cfg is nil")
	}

	return instance.readwriter
}

func ReadWriterLocation() string {
	if instance == nil {
		panic("cfg is nil")
	}

	return instance.readwriterLocation
}

func SheetsServiceAccountPath() string {
	if instance == nil {
		panic("cfg is nil")
	}

	return instance.sheetsServiceAccountKeyPath
}
