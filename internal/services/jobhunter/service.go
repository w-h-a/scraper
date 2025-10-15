package jobhunter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/w-h-a/scraper/internal/clients/reader"
	"github.com/w-h-a/scraper/internal/clients/readwriter"
	"github.com/w-h-a/scraper/internal/clients/scraper"
	"github.com/w-h-a/scraper/internal/jobpost"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	scraper    scraper.Scraper
	readwriter readwriter.ReadWriter
	tracer     trace.Tracer
}

func (s *Service) Start(ch chan struct{}) error {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.hunt()
	}()

	tick := time.NewTicker(24 * time.Hour)
	defer tick.Stop()

huntLoop:
	for {
		select {
		case <-ch:
			break huntLoop
		case <-tick.C:
			wg.Add(1)

			go func() {
				defer wg.Done()
				s.hunt()
			}()
		}
	}

	wg.Wait()

	return nil
}

func (s *Service) hunt() {
	ctx, span := s.tracer.Start(context.Background(), "JobHuntCycle")
	defer span.End()

	if err := s.ExecuteJobHunt(ctx); err != nil {
		fmt.Printf("FATAL: job hunt failed: %v", err)
		span.RecordError(err)
		return
	}

	fmt.Println("INFO: job hunt complete")
	span.AddEvent("JobHuntCompleted")
}

func (s *Service) ExecuteJobHunt(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "ExecuteJobHunt")
	defer span.End()

	span.AddEvent("JobHuntStarted")

	const linkReadRange = "A:D"

	existingLinks, err := s.readwriter.ReadExisting(ctx, reader.ReadExistingWithQuery(linkReadRange))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to read existing links: %s", err)
	}

	span.SetAttributes(attribute.Int("deduplication.set.size", len(existingLinks)))

	var wg sync.WaitGroup
	jobChan := make(chan jobpost.JobPost, 100)

	feedCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for source, url := range RSSFeeds {
		wg.Add(1)
		go s.processFeed(feedCtx, source, url, existingLinks, jobChan, &wg)
	}

	go func() {
		wg.Wait()
		close(jobChan)
		span.AddEvent("AllFeedsProcessed")
	}()

	var newJobs []jobpost.JobPost

	for job := range jobChan {
		newJobs = append(newJobs, job)
	}

	if len(newJobs) == 0 {
		span.AddEvent("NoNewJobsFound")
		return nil
	}

	span.SetAttributes(attribute.Int("jobs.newly_found", len(newJobs)))

	rowsToAppend := s.convertJobPostsToGenericRows(newJobs)

	return s.readwriter.WriteBatch(ctx, rowsToAppend)
}

func (s *Service) processFeed(ctx context.Context, sourceName string, url string, existingLinks map[string]bool, jobChan chan<- jobpost.JobPost, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, span := s.tracer.Start(ctx, "processFeed")
	defer span.End()

	span.SetAttributes(attribute.String("feed.source", sourceName))

	feed, err := s.scraper.Scrape(ctx, url)
	if err != nil {
		span.RecordError(err)
		return
	}

	newCount := 0

	for _, item := range feed.Items {
		if existingLinks[item.Link] {
			continue
		}

		dateString := ""

		if item.PublishedParsed != nil {
			dateString = item.PublishedParsed.In(time.Local).Format("2006-01-02 15:04:05")
		} else if item.UpdatedParsed != nil {
			dateString = item.UpdatedParsed.In(time.Local).Format("2006-01-02 15:04:05")
		} else {
			dateString = "N/A"
		}

		rawContent := item.Content

		if len(rawContent) == 0 {
			rawContent = item.Description
		}

		jobPost := jobpost.JobPost{
			DatePosted:     dateString,
			Source:         sourceName,
			JobTitle:       item.Title,
			Link:           item.Link,
			RawDescription: rawContent,
			Status:         "New",
		}

		jobChan <- jobPost

		newCount++
	}

	span.SetAttributes(attribute.Int("jobs.scraped_new", newCount))
	span.AddEvent("FeedProcessingFinished", trace.WithAttributes(attribute.Int("items.added", newCount)))
}

func (s *Service) convertJobPostsToGenericRows(jobs []jobpost.JobPost) [][]any {
	rows := make([][]any, len(jobs))

	for i, job := range jobs {
		rows[i] = []any{
			job.DatePosted,
			job.Source,
			job.JobTitle,
			job.Link,
			job.RawDescription,
			job.Status,
		}
	}

	return rows
}

func New(scraper scraper.Scraper, readwriter readwriter.ReadWriter) *Service {
	return &Service{
		scraper:    scraper,
		readwriter: readwriter,
		tracer:     otel.Tracer("job-hunter"),
	}
}
