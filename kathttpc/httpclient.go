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

func NewDefaultRetryableHttpClient(ctx context.Context) *http.Client {
	// This workaround allows the use of the default http client in tests (we can use httpmock)
	if katapp.RunningInTest(ctx) {
		return http.DefaultClient
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
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

type BodyRequest struct {
	Body                    io.Reader
	Headers                 http.Header
	ExpectSuccessStatusOnly bool
}

type BodyResponse struct {
	Body      []byte
	IsSuccess bool
	Response  *http.Response
}

type JsonRequest[TBody any] struct {
	Body                    *TBody
	Headers                 http.Header
	ExpectSuccessStatusOnly bool
}

type JsonResponse[TBody any, TErr any] struct {
	Body      TBody
	ErrBody   TErr
	IsSuccess bool
	Response  *http.Response
}

func DoBodyRequest(
	ctx context.Context, client *http.Client, method string, reqURL string, req BodyRequest,
) (*BodyResponse, error) {
	// Create a grouped logger
	logger := katapp.Logger(ctx).WithGroup("kathttpc.DoBodyRequest").With(
		"reqURL", reqURL,
		"method", method,
	)

	httpReq, err := http.NewRequestWithContext(ctx, method, reqURL, req.Body)
	if err != nil {
		logger.ErrorContext(ctx, "error creating request", "error", err)
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if req.Headers != nil {
		for k, h := range req.Headers {
			for _, v := range h {
				httpReq.Header.Add(k, v)
			}
		}
	}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		emsg := "error performing request"
		logger.ErrorContext(ctx, emsg, "error", err)
		return nil, fmt.Errorf("error %s: %w", emsg, err)
	}
	defer httpResp.Body.Close()

	resp := &BodyResponse{
		IsSuccess: httpResp.StatusCode >= 200 && httpResp.StatusCode < 300,
		Response:  httpResp,
	}
	if req.ExpectSuccessStatusOnly && !resp.IsSuccess {
		emsg := fmt.Sprintf("unexpected status code: %d", httpResp.StatusCode)
		logger.ErrorContext(ctx, emsg)
		return nil, &UnexpectedStatusCodeError{StatusCode: httpResp.StatusCode}
	}

	resp.Body, err = io.ReadAll(httpResp.Body)
	if err != nil {
		emsg := "failed to read response"
		logger.ErrorContext(ctx, emsg, "error", err)
		return nil, fmt.Errorf("%s: %w", emsg, err)
	}
	return resp, nil
}

func DoJsonRequest[TReqBody any, TRespBody any, TRespErr any](
	ctx context.Context, client *http.Client, method string, reqURL string, req JsonRequest[TReqBody],
) (*JsonResponse[TRespBody, TRespErr], error) {
	var body []byte
	var err error
	if req.Body != nil {
		body, err = json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}
	headers := withJsonHeaders(req.Headers)
	bodyResp, err := DoBodyRequest(
		ctx, client, method, reqURL, BodyRequest{
			Body:                    bytes.NewBuffer(body),
			Headers:                 headers,
			ExpectSuccessStatusOnly: req.ExpectSuccessStatusOnly,
		},
	)
	if err != nil {
		return nil, err
	}

	resp := &JsonResponse[TRespBody, TRespErr]{
		IsSuccess: bodyResp.IsSuccess,
		Response:  bodyResp.Response,
	}
	// If no content, return success
	if len(bodyResp.Body) == 0 {
		return resp, nil
	}
	if !bodyResp.IsSuccess {
		err = json.Unmarshal(bodyResp.Body, &resp.ErrBody)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal error response: %w", err)
		}
	} else {
		err = json.Unmarshal(bodyResp.Body, &resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}
	return resp, nil
}

func DoJsonGetRequest[TRespBody any, TRespErr any](
	ctx context.Context, client *http.Client, method string, reqURL string, req JsonRequest[any],
) (*JsonResponse[TRespBody, TRespErr], error) {
	return DoJsonRequest[any, TRespBody, TRespErr](ctx, client, method, reqURL, req)
}

func DoJsonPostRequest[TReqBody any, TRespBody any, TRespErr any](
	ctx context.Context, client *http.Client, method string, reqURL string, req JsonRequest[TReqBody],
) (*JsonResponse[TRespBody, TRespErr], error) {
	return DoJsonRequest[TReqBody, TRespBody, TRespErr](ctx, client, method, reqURL, req)
}

func DoJsonPutRequest[TReqBody any, TRespBody any, TRespErr any](
	ctx context.Context, client *http.Client, method string, reqURL string, req JsonRequest[TReqBody],
) (*JsonResponse[TRespBody, TRespErr], error) {
	return DoJsonRequest[TReqBody, TRespBody, TRespErr](ctx, client, method, reqURL, req)
}

func DoJsonDeleteRequest[TRespBody any, TRespErr any](
	ctx context.Context, client *http.Client, method string, reqURL string, req JsonRequest[any],
) (*JsonResponse[TRespBody, TRespErr], error) {
	return DoJsonRequest[any, TRespBody, TRespErr](ctx, client, method, reqURL, req)
}

func withJsonHeaders(headers http.Header) http.Header {
	reqHeaders := http.Header{}
	if headers != nil {
		maps.Copy(reqHeaders, headers)
	}
	reqHeaders["Content-Type"] = []string{"application/json"}
	return reqHeaders
}

func LocalURL(port int, path string) string {
	return fmt.Sprintf("http://0.0.0.0:%d/%s", port, strings.TrimPrefix(path, "/"))
}

func WaitForURLToBecomeReady(ctx context.Context, URL string) {
	logger := katapp.Logger(ctx).WithGroup("kathttpc.WaitForURLToBecomeReady").With("URL", URL)
	for i := 0; i < 30; i++ {
		logger.InfoContext(ctx, "probing URL")
		resp, err := http.Get(URL)
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
	panic(fmt.Sprintf("failed to become ready: %s", URL))
}
