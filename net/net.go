package net

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultRequestTimeout = 35 * time.Second

type Requester func(req *http.Request) ([]byte, error)

type Wrapper func(Requester) Requester

func Compose(wrappers ...Wrapper) Wrapper {
	return func(r Requester) Requester {
		for _, wrapper := range wrappers {
			r = wrapper(r)
		}
		return r
	}
}

func TokenAuth(token string) Wrapper {
	return func(r Requester) Requester {
		return func(req *http.Request) ([]byte, error) {
			req.Header.Add("Authorization", fmt.Sprintf("bearer %s", token))
			return r(req)
		}
	}
}

// computeJitter calcula um jitter temporal em milissegundos para evitar sincronizacao de retries
func computeJitter(maxJitter time.Duration) time.Duration {
	if maxJitter <= 0 {
		return 0
	}
	nanos := time.Now().UnixNano()
	if nanos < 0 {
		nanos = -nanos
	}
	return time.Duration(nanos % int64(maxJitter))
}

func MakeRequester(client *http.Client) Requester {
	return func(req *http.Request) ([]byte, error) {
		ctx, cancel := context.WithTimeout(req.Context(), DefaultRequestTimeout)
		defer cancel()

		reqWithCtx := req.WithContext(ctx)
		resp, err := client.Do(reqWithCtx)
		if err != nil {
			return []byte{}, err
		}
		defer resp.Body.Close()

		bodyText, err := io.ReadAll(resp.Body)
		if err != nil {
			return []byte{}, err
		}

		bodySnippet := string(bodyText)
		if len(bodySnippet) > 200 {
			bodySnippet = bodySnippet[:200] + "..."
		}

		// 1. Tratamento de Rate Limit: HTTP 429 ou HTTP 403 (Primary e Secondary Rate Limit)
		isSecondaryRateLimit := strings.Contains(string(bodyText), "secondary rate limit")
		isRateLimit := resp.StatusCode == http.StatusTooManyRequests ||
			(resp.StatusCode == http.StatusForbidden && (resp.Header.Get("X-RateLimit-Remaining") == "0" || isSecondaryRateLimit))

		if isRateLimit {
			// GitHub exige esperar pelo menos 60s quando atinge secondary rate limit
			backoff := 62*time.Second + computeJitter(5*time.Second)
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, err := time.ParseDuration(retryAfter + "s"); err == nil && seconds > 0 {
					backoff = seconds + (2 * time.Second) + computeJitter(3*time.Second)
				}
			}
			time.Sleep(backoff)
			return []byte{}, fmt.Errorf("github api rate limit exceeded (HTTP %d): %s", resp.StatusCode, bodySnippet)
		}

		// 2. Erros transitórios de infraestrutura da GitHub (502 Bad Gateway / 503 / 504 Gateway Timeout)
		if resp.StatusCode >= 500 {
			// Backoff adaptativo ampliado com jitter (15s a 20s) para permitir descarte de cache e alívio do gateway
			backoff := 15*time.Second + computeJitter(5*time.Second)
			time.Sleep(backoff)
			return []byte{}, fmt.Errorf("github server error (HTTP %d): %s", resp.StatusCode, bodySnippet)
		}

		// 3. Demais códigos não-OK (ex: 400, 401, 403 não rate-limit)
		if resp.StatusCode != http.StatusOK {
			return []byte{}, fmt.Errorf("unexpected http status %d: %s", resp.StatusCode, bodySnippet)
		}

		return bodyText, nil
	}
}
