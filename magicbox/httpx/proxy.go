package httpx

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
)

type ProxyClient struct {
	Host string
}

func (p *ProxyClient) Proxy(ctx context.Context, w http.ResponseWriter, r *http.Request, target string) error {
	query := r.URL.Query()

	api, err := url.JoinPath(p.Host, target)
	if err != nil {
		return err
	}
	toURL, err := url.Parse(api)
	if err != nil {
		return err
	}

	hx := NewClient(GetHTTPClient())

	headers := r.Header.Clone()
	if r.Host != "" {
		headers.Set("Host", r.Host)
	}
	headers.Set("X-Forwarded-Proto", "http")
	if r.TLS != nil {
		headers.Set("X-Forwarded-Proto", "https")
	}
	if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		headers.Set("X-Forwarded-For", clientIP)
	}

	opts := []Option{
		WithHeaders(headers),
		WithQuery(query),
	}
	if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
		opts = append(opts, WithBodyReader(r.Body))
	}

	resp, err := hx.Do(ctx, Method(r.Method), toURL.String(), opts...)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)

	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}

	_, err = io.Copy(w, resp.Body)
	return err
}
