package kattest

import (
	"context"
	"fmt"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/kathttpc"
	"net/http"
)

func LocalHttpJsonGetRequest[T any](
	ctx context.Context, cfg *katapp.ServerConfig, path string, headers map[string]string,
) (*T, http.Header, error) {
	URL := fmt.Sprintf("http://0.0.0.0:%d/%s", cfg.Port, path)
	return kathttpc.DoJsonGet[T](ctx, http.DefaultClient, URL, headers)
}

func LocalHttpJsonPostRequest[TReq any, TResp any](
	ctx context.Context, cfg *katapp.ServerConfig, path string, headers map[string]string, reqBody *TReq,
) (*TResp, http.Header, error) {
	URL := fmt.Sprintf("http://0.0.0.0:%d/%s", cfg.Port, path)
	return kathttpc.DoJsonPost[TReq, TResp](ctx, http.DefaultClient, URL, headers, reqBody)
}
