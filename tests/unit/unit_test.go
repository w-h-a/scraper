package unit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/require"
	mockreadwriter "github.com/w-h-a/scraper/internal/services/clients/readwriter/mock"
	mockscraper "github.com/w-h-a/scraper/internal/services/clients/scraper/mock"
	"github.com/w-h-a/scraper/internal/services/jobhunter"
)

func createMockFeed(count int) *gofeed.Feed {
	feed := &gofeed.Feed{
		Title: "Mock Feed",
		Items: make([]*gofeed.Item, count),
	}
	now := time.Now()
	for i := 0; i < count; i++ {
		link := fmt.Sprintf("http://joblink.com/%d", i)
		feed.Items[i] = &gofeed.Item{
			Title:           fmt.Sprintf("Job %d", i),
			Link:            link,
			Description:     "Test Description",
			PublishedParsed: &now,
		}
	}
	return feed
}

func TestJobHunter_ExecuteJobHunt_Success(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Log("SKIPPING UNIT TEST")
		return
	}

	ctx := context.Background()

	// 1. Arrange
	mockFeed := createMockFeed(3)

	mockReadWriter := mockreadwriter.NewReadWriter(
		mockreadwriter.WithExistingLinksKey(map[string]bool{}),
	)

	mockScraper := mockscraper.NewScraper(
		mockscraper.WithFeed(mockFeed),
	)

	service := jobhunter.New(mockScraper, mockReadWriter)

	// 2. Act
	err := service.ExecuteJobHunt(ctx)

	// 3. Assert
	require.NoError(t, err)

	// 2 feeds run concurrently.
	// Each feed processes 3 jobs.
	// Total jobs written = 3 + 3 = 6.
	expectedWritten := 6
	require.Equal(t, expectedWritten, len(mockReadWriter.RowsWritten))
}

func TestJobHunter_ExecuteJobHunt_Deduplication(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Log("SKIPPING UNIT TEST")
		return
	}

	ctx := context.Background()

	// 1. Arrange
	mockFeed := createMockFeed(5)

	mockReadWriter := mockreadwriter.NewReadWriter(
		mockreadwriter.WithExistingLinksKey(map[string]bool{
			mockFeed.Items[0].Link: true,
			mockFeed.Items[1].Link: true,
			mockFeed.Items[4].Link: true,
		}),
	)

	mockScraper := mockscraper.NewScraper(
		mockscraper.WithFeed(mockFeed),
	)

	service := jobhunter.New(mockScraper, mockReadWriter)

	// 2. Act
	err := service.ExecuteJobHunt(ctx)

	// 3. Assert
	require.NoError(t, err)

	// 2 feeds run concurrently.
	// Each feed processes 5 jobs.
	// 3 existing.
	// Total jobs written = 2 + 2 = 4.
	expectedWritten := 4
	require.Equal(t, expectedWritten, len(mockReadWriter.RowsWritten))
}

func TestJobHunter_ExecuteJobHunt_ScraperFails(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Log("SKIPPING UNIT TEST")
		return
	}

	ctx := context.Background()

	// 1. Arrange
	expectedErr := errors.New("network failed to fetch feed")

	mockReadWriter := mockreadwriter.NewReadWriter(
		mockreadwriter.WithExistingLinksKey(map[string]bool{}),
	)

	mockScraper := mockscraper.NewScraper(
		mockscraper.WithErr(expectedErr),
	)

	service := jobhunter.New(mockScraper, mockReadWriter)

	// 2. Act
	err := service.ExecuteJobHunt(ctx)

	// 3. Assert
	// handles gracefully
	require.NoError(t, err)
}

func TestJobHunter_ExecuteJobHunt_StoreReadFails(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Log("SKIPPING UNIT TEST")
		return
	}

	ctx := context.Background()

	// 1. Arrange
	expectedErr := errors.New("database connection failed")

	mockReadWriter := mockreadwriter.NewReadWriter(
		mockreadwriter.WithReadErr(expectedErr),
	)

	mockScraper := mockscraper.NewScraper(
		mockscraper.WithFeed(createMockFeed(1)),
	)

	service := jobhunter.New(mockScraper, mockReadWriter)

	// 2. Act
	err := service.ExecuteJobHunt(ctx)

	// 3. Assert
	require.Error(t, err)
	require.Contains(t, err.Error(), expectedErr.Error())
}
