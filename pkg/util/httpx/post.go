package httpx

import (
	"context"
	"net/http"
	"net/url"
)

func Post(ctx context.Context, api string, body []byte) (*http.Response, error) {
	return NewClient(WithContext(ctx)).Post(api, body)
}

func PostJson(ctx context.Context, api string, body []byte) (*http.Response, error) {
	return NewClient(WithContext(ctx)).PostJson(api, body)
}

func PostForm(ctx context.Context, api string, data url.Values) (*http.Response, error) {
	return NewClient(WithContext(ctx)).PostForm(api, data)
}

func PostJsonWithHeader(ctx context.Context, api string, body []byte, header http.Header) (*http.Response, error) {
	return NewClient(WithContext(ctx), WithHeader(header)).PostJson(api, body)
}

func PostJsonWithOptions(ctx context.Context, api string, body []byte, opts ...Option) (*http.Response, error) {
	return NewClient(append([]Option{WithContext(ctx)}, opts...)...).PostJson(api, body)
}
