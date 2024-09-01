package cmd

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
)

// otel:
//   meter:
//     service_name: oneway
//     exporter:
//       type: prometheus
//       prometheus:
//         endpoint: "http://prometheus:9090/metrics"

const Timeout = 10 * time.Second

type MeterConfig struct {
	Listen string `mapstructure:"listen"`
	Path   string `mapstructure:"path"`
}

type OtelConfig struct {
	Meter       *MeterConfig `mapstructure:"meter"`
	ServiceName string       `mapstructure:"service_name"`
}

func InitOtel(ctx context.Context, cfg *OtelConfig) error {
	var (
		shutdownFuncs []func(context.Context) error
		err           error
	)

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}

		shutdownFuncs = nil

		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up meter provider.
	meterProvider, err := newMeterProvider(cfg.Meter)
	if err != nil {
		handleErr(err)
		return err
	}

	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Serve metrics.
	go func() {
		err := serveMetrics(ctx, cfg.Meter)
		if err != nil {
			slog.Error("failed to serve metrics", slog.Any("error", err))
		}
	}()

	// Set up logger provider
	loggerProvider, err := newLoggerProvider(cfg)
	if err != nil {
		handleErr(err)
		return err
	}

	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)

	go func() {
		<-ctx.Done()

		if err := shutdown(ctx); err != nil {
			slog.Error("failed to shutdown otel", slog.Any("error", err))
		}
	}()

	return err
}

func newMeterProvider(_ *MeterConfig) (*metric.MeterProvider, error) {
	metricExporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metricExporter),
	)

	return meterProvider, nil
}

func newLoggerProvider(cfg *OtelConfig) (*log.LoggerProvider, error) {
	logExporter, err := stdoutlog.New()
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)

	slog.SetDefault(
		otelslog.NewLogger(
			cfg.ServiceName,
			otelslog.WithLoggerProvider(loggerProvider)))

	return loggerProvider, nil
}

func serveMetrics(ctx context.Context, cfg *MeterConfig) error {
	mux := http.NewServeMux()
	mux.Handle(cfg.Path, promhttp.Handler())

	httpSrv := &http.Server{
		Handler:      mux,
		ReadTimeout:  Timeout,
		WriteTimeout: Timeout,
	}

	slog.Info("serving metrics", slog.Any("listen", cfg.Listen), slog.Any("path", cfg.Path))

	lis, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()

		if err := httpSrv.Close(); err != nil {
			slog.Error("failed to close metric server", slog.Any("error", err))
		}
	}()

	return httpSrv.Serve(lis)
}
