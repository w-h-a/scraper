package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/w-h-a/scraper/internal/clients/readwriter"
	"github.com/w-h-a/scraper/internal/clients/readwriter/sheets"
	"github.com/w-h-a/scraper/internal/clients/scraper"
	"github.com/w-h-a/scraper/internal/clients/scraper/feed"
	"github.com/w-h-a/scraper/internal/config"
	"github.com/w-h-a/scraper/internal/services/jobhunter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func main() {
	ctx := context.Background()

	// config
	config.New()

	// setup tracer
	tracer, err := initTracer(ctx)
	if err != nil {
		panic(err)
	}
	defer tracer.Shutdown(ctx)

	// setup clients
	rw, err := initReadWriter(ctx)
	if err != nil {
		panic(err)
	}

	s, err := initScraper(ctx)
	if err != nil {
		panic(err)
	}

	// wait group & stop channels
	var wg sync.WaitGroup
	stopChannels := map[string]chan struct{}{}
	numServices := 0

	// job hunter
	hunter := jobhunter.New(s, rw)
	numServices += 1

	// error chan
	errCh := make(chan error, numServices)

	// start job hunter
	stop := make(chan struct{})
	stopChannels["hunter"] = stop

	wg.Add(1)
	go func() {
		defer wg.Done()
		errCh <- hunter.Start(stop)
	}()

	// block
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			fmt.Printf("service exited unexpectedly: %v", err)
		}
	case _ = <-signalCh:
		fmt.Printf("initiating graceful shutdown")
	}

	// graceful shutdown
	if stop, ok := stopChannels["hunter"]; ok {
		close(stop)
	}

	wait := make(chan struct{})
	go func() {
		defer close(wait)
		wg.Wait()
	}()

	select {
	case <-wait:
	case <-time.After(30 * time.Second):
	}
}

func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.Name()),
			semconv.DeploymentEnvironment(config.Env()),
			semconv.ServiceVersion(config.Version()),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource for tracer: %v", err)
	}

	exporterOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(config.TracesAddress()),
	}

	if len(config.TracesAPIKeyValue()) > 0 {
		exporterOpts = append(exporterOpts, otlptracehttp.WithHeaders(map[string]string{
			config.TracesAPIKeyHeader(): config.TracesAPIKeyValue(),
		}))
	} else {
		exporterOpts = append(exporterOpts, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, exporterOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter for traces: %v", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(
			sdktrace.NewBatchSpanProcessor(
				exporter,
			),
		),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider, nil
}

func initReadWriter(_ context.Context) (readwriter.ReadWriter, error) {
	return sheets.NewReadWriter(
		readwriter.WithLocation(config.ReadWriterLocation()),
		sheets.WithServiceAccountKeyPath(config.SheetsServiceAccountPath()),
	), nil
}

func initScraper(_ context.Context) (scraper.Scraper, error) {
	return feed.NewScraper(), nil
}
