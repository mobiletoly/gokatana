package kathttpc

import (
	"context"
	"github.com/mobiletoly/gokatana/katapp"
	"net/http"
)

func LocalHttpJsonGetRequest[TRespBody any](
	ctx context.Context, cfg *katapp.ServerConfig, path string, headers http.Header,
) (*TRespBody, http.Header, error) {
	resp, err := DoJsonGetRequest[TRespBody, any](
		ctx, http.DefaultClient, "GET", LocalURL(cfg.Port, path), JsonRequest[any]{
			ExpectSuccessStatusOnly: true,
			Headers:                 headers,
		},
	)
	if err != nil {
		return nil, nil, err
	}
	return &resp.Body, resp.Response.Header, nil
}

func LocalHttpJsonPostRequest[TReqBody any, TRespBody any](
	ctx context.Context, cfg *katapp.ServerConfig, path string, headers http.Header, reqBody *TReqBody,
) (*TRespBody, http.Header, error) {
	resp, err := DoJsonPostRequest[TReqBody, TRespBody, any](
		ctx, http.DefaultClient, "POST", LocalURL(cfg.Port, path), JsonRequest[TReqBody]{
			ExpectSuccessStatusOnly: true,
			Body:                    reqBody,
			Headers:                 headers,
		},
	)
	if err != nil {
		return nil, nil, err
	}
	return &resp.Body, resp.Response.Header, nil
}

func LocalHttpJsonPutRequest[TReqBody any, TRespBody any](
	ctx context.Context, cfg *katapp.ServerConfig, path string, headers http.Header, reqBody *TReqBody,
) (*TRespBody, http.Header, error) {
	resp, err := DoJsonPutRequest[TReqBody, TRespBody, any](
		ctx, http.DefaultClient, "PUT", LocalURL(cfg.Port, path), JsonRequest[TReqBody]{
			ExpectSuccessStatusOnly: true,
			Body:                    reqBody,
			Headers:                 headers,
		},
	)
	if err != nil {
		return nil, nil, err
	}
	return &resp.Body, resp.Response.Header, nil
}

func LocalHttpJsonDeleteRequest[TRespBody any](
	ctx context.Context, cfg *katapp.ServerConfig, path string, headers http.Header,
) (*TRespBody, http.Header, error) {
	resp, err := DoJsonDeleteRequest[TRespBody, any](
		ctx, http.DefaultClient, "DELETE", LocalURL(cfg.Port, path), JsonRequest[any]{
			ExpectSuccessStatusOnly: true,
			Headers:                 headers,
		},
	)
	if err != nil {
		return nil, nil, err
	}
	return &resp.Body, resp.Response.Header, nil
}
