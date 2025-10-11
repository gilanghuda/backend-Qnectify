package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	serviceName           = "axiom-go-otel"
	serviceVersion        = "0.1.0"
	deploymentEnvironment = "production"
)

var (
	httpClient   *http.Client
	ingestURL    string
	axiomToken   string
	axiomDataset string
)

func SetupTracer() (func(context.Context) error, error) {
	ctx := context.Background()
	if _, err := InstallExportPipeline(ctx); err != nil {
		return nil, err
	}
	return func(context.Context) error { return nil }, nil
}

func InstallExportPipeline(ctx context.Context) (func(context.Context) error, error) {
	token := os.Getenv("AXIOM_API_TOKEN")
	dataset := os.Getenv("AXIOM_DATASET")
	endpoint := os.Getenv("AXIOM_INGEST_ENDPOINT")

	if token == "" || dataset == "" {
		return nil, fmt.Errorf("AXIOM_API_TOKEN and AXIOM_DATASET environment variables must be set")
	}

	if endpoint == "" {
		endpoint = "https://api.axiom.co/v1/datasets/" + url.PathEscape(dataset) + "/ingest"
	}

	if u, err := url.Parse(endpoint); err != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid AXIOM_INGEST_ENDPOINT: %s", endpoint)
	}

	axiomToken = token
	axiomDataset = dataset
	ingestURL = endpoint

	httpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{},
		},
	}

	log.Info().Str("ingest_url", ingestURL).Str("dataset", axiomDataset).Msg("axiom ingest configured")

	return func(context.Context) error { return nil }, nil
}

func EmitLog(ctx context.Context, level string, message string, attrs map[string]string) error {
	if httpClient == nil || ingestURL == "" {
		return fmt.Errorf("ingest not configured")
	}

	event := map[string]interface{}{
		"_time":   time.Now().UTC().Format(time.RFC3339),
		"message": message,
		"level":   level,
		"service": serviceName,
	}
	for k, v := range attrs {
		event[k] = v
	}

	payload, err := json.Marshal([]map[string]interface{}{event})
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal ingest payload")
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ingestURL, bytes.NewReader(payload))
	if err != nil {
		log.Error().Err(err).Msg("failed to create ingest request")
		return err
	}
	req.Header.Set("Authorization", "Bearer "+axiomToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("failed to send ingest request")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error().Int("status", resp.StatusCode).Msg("ingest request failed")
		return fmt.Errorf("ingest request failed with status %d", resp.StatusCode)
	}

	return nil
}

func init() {

	EmitLogFunc = EmitLog
}
