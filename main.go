package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/w-h-a/scraper/internal/clients/readwriter"
	"github.com/w-h-a/scraper/internal/clients/readwriter/sheets"
	"github.com/w-h-a/scraper/internal/clients/scraper"
	"github.com/w-h-a/scraper/internal/clients/scraper/feed"
	"github.com/w-h-a/scraper/internal/config"
	"github.com/w-h-a/scraper/internal/services/jobhunter"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	globallog "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func main() {
	ctx := context.Background()

	// config
	config.New()

	// setup resource
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.Name()),
			semconv.DeploymentEnvironment(config.Env()),
			semconv.ServiceVersion(config.Version()),
		),
	)
	if err != nil {
		panic(err)
	}

	// setup logger
	lp, err := initLogger(ctx, res)
	if err != nil {
		panic(err)
	}
	defer lp.Shutdown(ctx)

	logger := otelslog.NewLogger(
		config.Name(),
		otelslog.WithLoggerProvider(lp),
	)

	slog.SetDefault(logger)

	// setup tp
	tp, err := initTracer(ctx, res)
	if err != nil {
		panic(err)
	}
	defer tp.Shutdown(ctx)

	// wait group & stop channels
	var wg sync.WaitGroup
	stopChannels := map[string]chan struct{}{}

	// setup
	rw, err := initReadWriter(ctx)
	if err != nil {
		panic(err)
	}

	s, err := initScraper(ctx)
	if err != nil {
		panic(err)
	}

	hunter := jobhunter.New(s, rw)
	stopChannels["hunter"] = make(chan struct{})

	// error and sig chans
	errCh := make(chan error, len(stopChannels))
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// start
	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.InfoContext(ctx, "starting job hunter service")
		errCh <- hunter.Run(stopChannels["hunter"])
	}()

	// block until shutdown
	select {
	case err := <-errCh:
		if err != nil {
			slog.ErrorContext(ctx, "service exited unexpectedly", "error", err)
			panic(err)
		}
	case _ = <-signalCh:
		slog.InfoContext(ctx, "initiating graceful shutdown")
		for _, stop := range stopChannels {
			close(stop)
		}
	}

	wg.Wait()

	close(errCh)

	for err := range errCh {
		if err != nil {
			slog.ErrorContext(ctx, "error upon shutting down", "error", err)
		}
	}

	slog.InfoContext(ctx, "shutdown")
}

func initLogger(ctx context.Context, res *resource.Resource) (*sdklog.LoggerProvider, error) {
	var exporter sdklog.Exporter
	var err error

	if len(config.LogsAPIKeyValue()) > 0 {
		exporter, err = otlploghttp.New(
			ctx,
			otlploghttp.WithEndpoint(config.LogsAddress()),
			otlploghttp.WithHeaders(map[string]string{
				config.LogsAPIKeyHeader(): config.LogsAPIKeyValue(),
			}),
		)
	} else {
		exporter, err = stdoutlog.New()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create exporter for logs: %v", err)
	}

	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(
			sdklog.NewBatchProcessor(exporter),
		),
	)

	globallog.SetLoggerProvider(loggerProvider)

	return loggerProvider, nil
}

func initTracer(ctx context.Context, res *resource.Resource) (*sdktrace.TracerProvider, error) {
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
