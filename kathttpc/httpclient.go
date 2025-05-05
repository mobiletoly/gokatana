package kathttpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/mobiletoly/gokatana/katapp"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"strings"
	"time"
)

type UnexpectedStatusCodeError struct {
	StatusCode int
}

func (e UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("unexpected status code: %d", e.StatusCode)
}

func NewRetryableHttpClient(ctx context.Context) *http.Client {
	// This workaround allows the use of the default http client in tests (we can use httpmock)
	if katapp.RunningInTest(ctx) {
		return http.DefaultClient
	}

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = &retryableLogger{logger: katapp.Logger(ctx).Logger}
	client := retryClient.StandardClient()
	return client
}

type retryableLogger struct {
	logger *slog.Logger
}

func (l *retryableLogger) Printf(format string, v ...interface{}) {
	l.logger.Debug(format, v...)
}

func DoRequest(
	ctx context.Context, client *http.Client, method string, reqURL string, headers map[string]string, body io.Reader,
) ([]byte, http.Header, error) {
	// Create a grouped logger
	logger := katapp.Logger(ctx).WithGroup("kathttpc.DoRequest").With(
		"reqURL", reqURL,
		"method", method,
	)

	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		logger.ErrorContext(ctx, "error creating request", "error", err)
		return nil, nil, fmt.Errorf("error creating request: %w", err)
	}

	if headers != nil {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.ErrorContext(ctx, "error performing request", "error", err)
		return nil, nil, fmt.Errorf("error performing request : %w", err)
	}
	defer resp.Body.Close()

	header := resp.Header

	// Check the HTTP response status code
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		logger.ErrorContext(ctx, "unexpected status code", "status", resp.StatusCode)
		return nil, header, &UnexpectedStatusCodeError{StatusCode: resp.StatusCode}
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorContext(ctx, "failed to read response", "error", err)
		return nil, header, fmt.Errorf("failed to read response: %w", err)
	}
	return buf, header, nil
}

func DoJsonGet[T any](
	ctx context.Context, client *http.Client, reqURL string, headers map[string]string,
) (*T, http.Header, error) {
	logger := katapp.Logger(ctx).WithGroup("kathttpc.DoJsonGet").With(
		"reqURL", reqURL,
		"method", "GET",
	)
	reqHeaders := withJsonHeaders(headers)
	buf, header, err := DoRequest(ctx, client, "GET", reqURL, reqHeaders, nil)
	if err != nil {
		return nil, header, err
	}
	var result T
	err = json.Unmarshal(buf, &result)
	if err != nil {
		logger.ErrorContext(ctx, "failed to unmarshal content", "error", err)
		return nil, header, fmt.Errorf("failed to unmarshal content: %w", err)
	}
	return &result, header, nil
}

func DoJsonPost[TReq any, TResp any](
	ctx context.Context, client *http.Client, reqURL string, headers map[string]string, reqBody *TReq,
) (*TResp, http.Header, error) {
	logger := katapp.Logger(ctx).WithGroup("kathttpc.DoJsonPost").With(
		"reqURL", reqURL,
		"method", "POST",
	)
	body, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("failed to marshal request body", "error", err)
		return nil, nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	reqHeaders := withJsonHeaders(headers)
	buf, header, err := DoRequest(ctx, client, "POST", reqURL, reqHeaders, bytes.NewBuffer(body))
	if err != nil {
		return nil, header, err
	}
	var result TResp
	err = json.Unmarshal(buf, &result)
	if err != nil {
		logger.ErrorContext(ctx, "failed to unmarshal content", "error", err)
		return nil, header, fmt.Errorf("failed to unmarshal content: %w", err)
	}
	return &result, header, nil
}

func withJsonHeaders(headers map[string]string) map[string]string {
	reqHeaders := map[string]string{}
	if headers != nil {
		maps.Copy(reqHeaders, headers)
	}
	reqHeaders["Content-Type"] = "application/json"
	return reqHeaders
}

func WaitForURLToBecomeReady(ctx context.Context, conf *katapp.ServerConfig, path string) {
	reqURL := fmt.Sprintf("http://0.0.0.0:%d/%s", conf.Port, strings.TrimPrefix(path, "/"))
	logger := katapp.Logger(ctx).WithGroup("kathttpc.WaitForURLToBecomeReady").With("reqURL", reqURL)
	for i := 0; i < 30; i++ {
		logger.InfoContext(ctx, "probing URL")
		resp, err := http.Get(reqURL)
		if err != nil {
			logger.InfoContext(ctx, "URL is not ready yet")
		} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			logger.InfoContext(ctx, "URL is ready")
			return
		} else {
			logger.InfoContext(ctx, "URL is not ready yet", "status", resp.StatusCode)
		}
		time.Sleep(1 * time.Second)
	}
	panic(fmt.Sprintf("failed to become ready: %s", reqURL))
}
