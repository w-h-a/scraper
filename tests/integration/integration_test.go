package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/w-h-a/scraper/internal/clients/reader"
	"github.com/w-h-a/scraper/internal/clients/readwriter"
	"github.com/w-h-a/scraper/internal/clients/readwriter/sheets"
	"github.com/w-h-a/scraper/internal/clients/scraper/feed"
	"github.com/w-h-a/scraper/internal/services/jobhunter"
)

func TestIntegration_ReadWriter_CanReadAndWrite(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) == 0 {
		t.Log("SKIPPING INTEGRATION TEST")
		return
	}

	ctx := context.Background()

	rw := sheets.NewReadWriter(
		readwriter.WithLocation(os.Getenv("READ_WRITER_LOCATION")),
		sheets.WithServiceAccountKeyPath(os.Getenv("SHEETS_SERVICE_ACCOUNT_KEY_PATH")),
	)

	rw.ClearBatch(ctx)

	uniqueID := fmt.Sprintf("http://test.link.com/%d", time.Now().UnixNano())

	testRow := [][]any{
		{time.Now().Format("2006-01-02 15:04:05"), "TEST_INTEGRATION", "TEST JOB TITLE", uniqueID, "Test Description", "TEST"},
	}

	t.Run("AppendData", func(t *testing.T) {
		err := rw.WriteBatch(ctx, testRow)
		require.NoError(t, err)
	})

	t.Run("ReadExistingIDs", func(t *testing.T) {
		const testRange = "A:D"

		links, err := rw.ReadExisting(
			ctx,
			reader.ReadExistingWithQuery(testRange),
		)

		require.NoError(t, err)
		require.True(t, links[uniqueID])
	})
}

func TestSystem_FullJobHuntCycle(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) == 0 {
		t.Log("SKIPPING INTEGRATION TEST")
		return
	}

	ctx := context.Background()

	realScraper := feed.NewScraper()

	realReadWriter := sheets.NewReadWriter(
		readwriter.WithLocation(os.Getenv("READ_WRITER_LOCATION")),
		sheets.WithServiceAccountKeyPath(os.Getenv("SHEETS_SERVICE_ACCOUNT_KEY_PATH")),
	)

	realReadWriter.ClearBatch(ctx)

	service := jobhunter.New(realScraper, realReadWriter)

	// 2. Act
	err := service.ExecuteJobHunt(ctx)

	// 3. Assert:
	require.NoError(t, err)

	const checkRange = "A:D"

	finalLinks, err := realReadWriter.ReadExisting(
		ctx,
		reader.ReadExistingWithQuery(checkRange),
	)

	require.NoError(t, err)
	require.Greater(t, len(finalLinks), 0)
}
