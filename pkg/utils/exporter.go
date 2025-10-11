package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const (
	serviceName           = "axiom-go-otel"
	serviceVersion        = "0.1.0"
	deploymentEnvironment = "production"
)

func SetupTracer() (func(context.Context) error, error) {
	ctx := context.Background()
	return InstallExportPipeline(ctx)
}

func Resource() *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(serviceVersion),
		attribute.String("environment", deploymentEnvironment),
	)
}

func InstallExportPipeline(ctx context.Context) (func(context.Context) error, error) {
	token := os.Getenv("AXIOM_API_TOKEN")
	dataset := os.Getenv("AXIOM_DATASET")
	endpoint := os.Getenv("AXIOM_OTLP_ENDPOINT")
	path := os.Getenv("AXIOM_OTLP_PATH")
	orgID := os.Getenv("AXIOM_ORG_ID")

	if token == "" || dataset == "" || endpoint == "" {
		return nil, fmt.Errorf("AXIOM_API_TOKEN, AXIOM_DATASET, and AXIOM_OTLP_ENDPOINT environment variables must be set")
	}

	h := endpoint
	if strings.HasPrefix(h, "http://") || strings.HasPrefix(h, "https://") {
		u, err := url.Parse(h)
		if err != nil {
			return nil, fmt.Errorf("invalid AXIOM_OTLP_ENDPOINT: %w", err)
		}
		if u.Host == "" {
			return nil, fmt.Errorf("invalid AXIOM_OTLP_ENDPOINT, missing host")
		}
		h = u.Host
		if path == "" && u.Path != "" {
			path = u.Path
		}
	}

	if !strings.Contains(h, ":") {
		h = h + ":443"
	}

	if path == "" {
		path = "/v1/otlp/traces"
	}

	headers := map[string]string{
		"Authorization":   "Bearer " + token,
		"X-AXIOM-DATASET": dataset,
	}
	if orgID != "" {
		headers["X-AXIOM-ORG-ID"] = orgID
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(h),
		otlptracehttp.WithURLPath(path),
		otlptracehttp.WithHeaders(headers),
		otlptracehttp.WithTLSClientConfig(&tls.Config{}),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to create otlp http exporter")
		return nil, err
	}

	log.Info().Str("endpoint", h).Str("url_path", path).Str("dataset", dataset).Msg("otel exporter configured")

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(Resource()),
	)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tr := otel.Tracer("axiom-debug")
		_, span := tr.Start(ctx, "axiom-startup-test")
		span.End()

		if err := tracerProvider.ForceFlush(ctx); err != nil {
			log.Error().Err(err).Msg("failed to flush initial test span to axiom")
		} else {
			log.Info().Str("dataset", dataset).Msg("initial test span flushed to axiom")
		}
	}()

	return tracerProvider.Shutdown, nil
}
