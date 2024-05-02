package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	otel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otelMetrics "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

var (
	collectorEndpoint = flag.String("endpointDomain", "localhost:8429", "host:port")
	collectorURL      = flag.String("ingestPath", "/opentelemetry/api/v1/push", "url path for ingestion path")
	isSecure          = flag.Bool("isSecure", false, "enables https connection for metrics push")
	pushInterval      = flag.Duration("pushInterval", 1*time.Second, "how often push samples, aka scrapeInterval at pull model")
	jobName           = flag.String("metrics.jobName", "otlp", "job name for web-application")
	instanceName      = flag.String("metrics.instance", "localhost", "hostname of web-application instance")
)

func main() {
	flag.Parse()
	log.Printf("Starting web server...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Registering all handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/api/fast", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(`fast ok`))
	})
	mux.HandleFunc("/api/slow", func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(time.Second * 2)
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(`slow ok`))
	})

	// Wrapping handlers around middleware
	mw, err := newMetricsMiddleware(ctx, mux)
	if err != nil {
		panic(fmt.Sprintf("cannot build metricMiddleWare: %q", err))
	}

	// To handle shutdown signal
	mustStop := make(chan os.Signal, 1)
	signal.Notify(mustStop, os.Interrupt, syscall.SIGTERM)
	go func() {
		http.ListenAndServe(":8081", mw)
	}()
	log.Printf("web server started at localhost:8081.")

	<-mustStop
	log.Println("receive shutdown signal, stopping webserver")
	if err := mw.onShutdown(ctx); err != nil {
		log.Println("cannot shutdown metric provider ", err)
	}

	cancel()
	log.Printf("Done!")
}

func newMetricsProvider(ctx context.Context) (*metric.MeterProvider, *metric.PeriodicReader, error) {
	options := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(*collectorEndpoint),
		otlpmetrichttp.WithURLPath(*collectorURL),
	}

	if !*isSecure {
		options = append(options, otlpmetrichttp.WithInsecure())
	}

	metricExporter, err := otlpmetrichttp.New(ctx, options...)
	//options := []otlpmetricgrpc.Option{
	//	otlpmetricgrpc.WithEndpoint(*collectorEndpoint),
	//	otlpmetricgrpc.WithInsecure(),
	//}
	//if !*isSecure {
	//	options = append(options, otlpmetricgrpc.WithInsecure())
	//}
	//metricExporter, err := otlpmetricgrpc.New(ctx, options...)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create otlphttp exporter: %w", err)
	}

	// reader := metric.NewManualReader()
	reader := metric.NewPeriodicReader(metricExporter, metric.WithInterval(*pushInterval))

	resourceConfig, err := resource.New(ctx, resource.WithAttributes(attribute.String("service.name", "myapp"), attribute.String("job", *jobName), attribute.String("instance", *instanceName)))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create meter resource: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(resourceConfig),
		metric.WithReader(reader),
	)

	return meterProvider, reader, nil
}

func newMetricsMiddleware(ctx context.Context, h http.Handler) (*metricMiddleWare, error) {
	mw := &metricMiddleWare{
		ctx: ctx,
		h:   h,
	}
	mc, reader, err := newMetricsProvider(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot build metrics collector: %w", err)
	}

	otel.SetMeterProvider(mc)

	meter := mc.Meter("http")

	mw.requestsCount, err = meter.Int64Counter("http_requests_total")
	if err != nil {
		return nil, fmt.Errorf("cannot create http_requests_count counter: %w", err)
	}

	mw.activeRequests, err = meter.Int64UpDownCounter("http_requests_active")
	if err != nil {
		return nil, fmt.Errorf("cannot create http_requests_active counter: %w", err)
	}

	mw.requestsLatency, err = meter.Float64Histogram("http_request_duration_seconds", otelMetrics.WithDescription("The HTTP request latencies in seconds."))
	if err != nil {
		return nil, fmt.Errorf("cannot create http_request_duration_seconds histogram: %w", err)
	}

	mw.onShutdown = func(ctx context.Context) error {
		if err := mc.Shutdown(ctx); err != nil {
			return fmt.Errorf("cannot stop metric provider: %w", err)
		}
		return nil
	}

	mw.reader = reader

	return mw, nil
}

type metricMiddleWare struct {
	ctx             context.Context
	h               http.Handler
	requestsCount   otelMetrics.Int64Counter
	requestsLatency otelMetrics.Float64Histogram
	activeRequests  otelMetrics.Int64UpDownCounter
	reader          *metric.PeriodicReader

	onShutdown func(ctx context.Context) error
}

func (m *metricMiddleWare) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	path := r.URL.Path
	m.requestsCount.Add(m.ctx, 1, otelMetrics.WithAttributes(
		attribute.String("path", path)),
	)

	m.activeRequests.Add(m.ctx, 1, otelMetrics.WithAttributes(
		attribute.String("path", path)),
	)

	defer func() {
		m.activeRequests.Add(m.ctx, -1, otelMetrics.WithAttributes(
			attribute.String("path", path)),
		)

		m.requestsLatency.Record(m.ctx, time.Since(t).Seconds(), otelMetrics.WithAttributes(attribute.String("path", path)))
	}()

	// collectedMetrics := &metricdata.ResourceMetrics{}
	// m.reader.Collect(context.TODO(), collectedMetrics)

	// fmt.Printf("Collected metrics: %v\n", collectedMetrics)

	m.h.ServeHTTP(w, r)
}
