package riot

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Omashu-Data/api-data/riot-petitions/pkg/core"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
)

const (
	HeaderRateLimitApplication      = "x-app-rate-limit"
	HeaderRateLimitApplicationCount = "x-app-rate-limit-count"
	HeaderRateLimitMethod           = "x-method-rate-limit"
	HeaderRateLimitMethodCount      = "x-method-rate-limit-count"
	HeaderRateLimitType             = "x-rate-limit-type"
	HeaderRetryAfter                = "retry-after"
)

type RateLimitedClient struct {
	c             *retryablehttp.Client
	mu            sync.Mutex             // Mutex for application-level rate limiting
	methodMutexes map[string]*sync.Mutex // Mutexes for method-level rate limiting
}

// NewRateLimitedClient initializes a new RateLimitedClient with the given http.Client and RateLimiter.
func NewRateLimitedClient(application string, methods []core.Routes) (*RateLimitedClient, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 30 * time.Second
	retryClient.Logger = nil // Disable default logger

	// Create method-specific mutexes
	methodMutexes := make(map[string]*sync.Mutex)
	for _, method := range methods {
		methodMutexes[string(method)] = &sync.Mutex{}
	}

	// Use the custom backoff function
	retryClient.Backoff = customBackoff

	return &RateLimitedClient{c: retryClient, methodMutexes: methodMutexes}, nil
}

func (rlc *RateLimitedClient) Do(req *http.Request, method string) (*http.Response, error) {
	retryReq, err := retryablehttp.FromRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error creating retryable request: %w", err)
	}

	retryReq.Header.Set("X-Custom-Method", method) // Store method for use in CheckRetry

	rlc.c.CheckRetry = rlc.customCheckRetry

	startTime := time.Now()
	resp, err := rlc.c.Do(retryReq)
	if err != nil {
		return nil, err
	}
	duration := time.Since(startTime)

	log.Info().
		Str("method", req.Method).
		Str("url", req.URL.String()).
		Int("status", resp.StatusCode).
		Dur("duration", duration).
		Msg("Request completed")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading response body")
		return resp, err
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	return resp, nil
}

func (rlc *RateLimitedClient) customCheckRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if err != nil {
		return true, err
	}

	// Check for rate limit response
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get(HeaderRetryAfter)
		if retryAfter != "" {
			log.Info().Str("retryAfter", retryAfter).Msg("Rate limited, retrying after specified time")
			seconds, err := strconv.Atoi(retryAfter)
			if err == nil {

				log.Info().
					Int("retry_after_seconds", seconds).
					Dur("sleep_duration", time.Duration(seconds)*time.Second).
					Msgf("Rate limited, retrying after %d seconds", seconds)

				// Now, check if the rate limit is app-wide or method-specific
				rateLimitType := resp.Header.Get(HeaderRateLimitType)

				if rateLimitType == "application" {
					// Block all requests (application-wide rate limit)
					rlc.mu.Lock()
					defer rlc.mu.Unlock()
					time.Sleep(time.Duration(seconds) * time.Second)
					log.Info().Msg("Finished Sleeping - App-wide block")
				} else if rateLimitType == "method" {
					// Block only the specific method (method-level rate limit)
					method := resp.Request.Header.Get("X-Custom-Method")
					methodMutex, exists := rlc.methodMutexes[method]
					if exists {
						methodMutex.Lock()
						defer methodMutex.Unlock()
						time.Sleep(time.Duration(seconds) * time.Second)

						// Double-check the app-wide lock after method rate-limit sleep
						rlc.mu.Lock()
						defer rlc.mu.Unlock()
						log.Info().Msgf("Finished Sleeping - Method block for %s", method)
					}
				} else {
					log.Error().Msg("Invalid method, doing nothing")
				}
				return true, nil
			}
		}
	}

	if resp.StatusCode >= 500 || resp.StatusCode == 408 {
		// Retry for 5xx errors and request timeout (408)
		return true, nil
	}

	return false, nil
}

func customBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	// If we receive a 429, we don't want any additional backoff, so return 0 duration.
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		return 0 // No backoff for 429, since it's handled by the custom retry logic
	}
	// For other status codes, apply the default exponential backoff
	return retryablehttp.DefaultBackoff(min, max, attemptNum, resp)
}
