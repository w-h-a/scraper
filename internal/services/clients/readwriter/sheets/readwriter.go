package sheets

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/w-h-a/scraper/internal/services/clients/reader"
	"github.com/w-h-a/scraper/internal/services/clients/readwriter"
	"github.com/w-h-a/scraper/internal/services/clients/writer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type sheetsReadWriter struct {
	options readwriter.Options
	client  *sheets.Service
	tracer  trace.Tracer
}

func (s *sheetsReadWriter) ReadExisting(ctx context.Context, opts ...reader.ReadExistingOption) (map[string]bool, error) {
	_, span := s.tracer.Start(ctx, "sheets.ReadExisting")
	defer span.End()

	options := reader.NewReadExistingOptions(opts...)

	fullRange := "Sheet1" + "!" + options.Query
	span.SetAttributes(attribute.String("db.operation", "read_links"))

	rsp, err := s.client.Spreadsheets.Values.Get(s.options.Location, fullRange).Context(ctx).Do()
	if err != nil {
		if strings.Contains(err.Error(), "Unable to parse range") {
			span.AddEvent("SheetEmpty", trace.WithAttributes(attribute.String("warning", "sheet range was empty")))
			return map[string]bool{}, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to retrieve data from sheet: %w", err)
	}

	existingLinks := map[string]bool{}

	const linkColIndex = 3

	for i, row := range rsp.Values {
		if i == 0 || len(row) <= linkColIndex {
			continue
		}
		link := fmt.Sprintf("%v", row[linkColIndex])
		existingLinks[link] = true
	}

	span.SetAttributes(attribute.Int("deduplication.count", len(existingLinks)))

	return existingLinks, nil
}

func (s *sheetsReadWriter) WriteBatch(ctx context.Context, rows [][]any, _ ...writer.WriteBatchOption) error {
	_, span := s.tracer.Start(ctx, "sheets.WriteBatch")
	defer span.End()

	if len(rows) == 0 {
		return nil
	}

	span.SetAttributes(attribute.Int("rows.count", len(rows)))
	span.SetAttributes(attribute.String("db.operation", "append_data"))

	var valueRange sheets.ValueRange

	valueRange.Values = rows

	if _, err := s.client.Spreadsheets.Values.Append(s.options.Location, "Sheet1"+"!"+"A:H", &valueRange).Context(ctx).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to append data to sheet: %w", err)
	}

	span.AddEvent("DataSuccessfullyAppended", trace.WithAttributes(attribute.Int("records.written", len(rows))))

	return nil
}

func (s *sheetsReadWriter) ClearBatch(ctx context.Context, opts ...writer.ClearBatchOption) error {
	rsp, err := s.client.Spreadsheets.Get(s.options.Location).Context(ctx).Fields("sheets.properties").Do()
	if err != nil {
		return err
	}

	if len(rsp.Sheets) == 0 {
		return nil
	}

	sheetProperties := rsp.Sheets[0].Properties

	if sheetProperties.GridProperties.RowCount <= 1 {
		return nil
	}

	deleteRequest := sheets.Request{
		DeleteDimension: &sheets.DeleteDimensionRequest{
			Range: &sheets.DimensionRange{
				SheetId:    sheetProperties.SheetId,
				Dimension:  "ROWS",
				StartIndex: 1, // Start deleting from the second row (after headers)
				EndIndex:   sheetProperties.GridProperties.RowCount,
			},
		},
	}

	batchUpdateRequest := sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{&deleteRequest},
	}

	_, err = s.client.Spreadsheets.BatchUpdate(s.options.Location, &batchUpdateRequest).Context(ctx).Do()

	return err
}

func (rw *sheetsReadWriter) configure(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		detail := fmt.Sprintf("failed to read service account key: %v", err)
		panic(detail)
	}

	config, err := google.JWTConfigFromJSON(data, sheets.SpreadsheetsScope)
	if err != nil {
		detail := fmt.Sprintf("failed to parse service account key: %v", err)
		panic(detail)
	}

	ctx := context.Background()

	client := config.Client(ctx)

	sheetsClient, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		detail := fmt.Sprintf("failed to retrieve sheets client: %v", err)
		panic(detail)
	}

	rw.client = sheetsClient
}

func NewReadWriter(opts ...readwriter.Option) readwriter.ReadWriter {
	options := readwriter.NewOptions(opts...)

	rw := &sheetsReadWriter{
		options: options,
		tracer:  otel.Tracer("sheets-readwriter"),
	}

	if path, ok := getServiceAccountKeyPathFromCtx(options.Context); ok {
		rw.configure(path)
	}

	return rw
}
