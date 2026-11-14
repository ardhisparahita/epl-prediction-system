package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ardhisparahita/epl-prediction-system/backend/internal/dto"
)

const (
	predictionEndpoint = "/api/v1/predictions"
	maxResponseSize    = 1 << 20 // 1 MB
)

// MLClient handles communication with the Python ML Service.
type MLClient struct {
	baseURL    string
	httpClient *http.Client
}

// HTTPError represents a non-successful response from ML Service.
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf(
		"ML service returned HTTP %d: %s",
		e.StatusCode,
		e.Body,
	)
}

// NewMLClient creates a new ML Service HTTP client.
func NewMLClient(
	baseURL string,
	timeout time.Duration,
) (*MLClient, error) {
	baseURL = strings.TrimSpace(baseURL)

	if baseURL == "" {
		return nil, fmt.Errorf("ML service URL is required")
	}

	if timeout <= 0 {
		return nil, fmt.Errorf("ML service timeout must be positive")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid ML service URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("ML service URL must use HTTP or HTTPS")
	}

	if parsedURL.Host == "" {
		return nil, fmt.Errorf("ML service URL must contain a host")
	}

	if parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		return nil, fmt.Errorf(
			"ML service URL must not contain query parameters or fragments",
		)
	}

	return &MLClient{
		baseURL: strings.TrimRight(parsedURL.String(), "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Predict sends a prediction request to the Python ML Service.
func (c *MLClient) Predict(
	ctx context.Context,
	payload dto.PredictionRequest,
) (dto.PredictionResponse, error) {
	var result dto.PredictionResponse

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return result, fmt.Errorf(
			"failed to encode prediction request: %w",
			err,
		)
	}

	endpoint := c.baseURL + predictionEndpoint

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return result, fmt.Errorf(
			"failed to create ML service request: %w",
			err,
		)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return result, fmt.Errorf(
			"failed to communicate with ML service: %w",
			err,
		)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(
		io.LimitReader(resp.Body, maxResponseSize+1),
	)
	if err != nil {
		return result, fmt.Errorf(
			"failed to read ML service response: %w",
			err,
		)
	}

	if len(responseBody) > maxResponseSize {
		return result, fmt.Errorf(
			"ML service response exceeds maximum allowed size",
		)
	}

	if resp.StatusCode < http.StatusOK ||
		resp.StatusCode >= http.StatusMultipleChoices {

		message := strings.TrimSpace(string(responseBody))

		var apiError dto.MLServiceError

		if err := json.Unmarshal(responseBody, &apiError); err == nil &&
			apiError.Detail != nil {

			if detail, ok := apiError.Detail.(string); ok {
				message = detail
			} else if detail, err := json.Marshal(apiError.Detail); err == nil {
				message = string(detail)
			}
		}

		if message == "" {
			message = http.StatusText(resp.StatusCode)
		}

		return result, &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       message,
		}
	}

	if err := json.Unmarshal(responseBody, &result); err != nil {
		return result, fmt.Errorf(
			"failed to decode ML service response: %w",
			err,
		)
	}

	return result, nil
}
