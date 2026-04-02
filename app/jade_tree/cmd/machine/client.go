package machine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aide-family/magicbox/strutil/cnst"
	"google.golang.org/protobuf/encoding/protojson"

	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
)

const defaultEndpoint = "localhost:8004"

type apiClient struct {
	client    *http.Client
	authToken string
}

func newAPIClient(httpClient *http.Client, authToken string) *apiClient {
	return &apiClient{client: httpClient, authToken: strings.TrimSpace(authToken)}
}

func normalizeBaseURL(endpoint string) (string, error) {
	base := strings.TrimSpace(endpoint)
	if base == "" {
		return "", fmt.Errorf("endpoint is empty")
	}
	if !strings.Contains(base, "://") {
		base = "http://" + base
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Host == "" {
		u.Host = defaultEndpoint
	}
	return u.String(), nil
}

func (c *apiClient) getMachineInfo(ctx context.Context, endpoint string) (*apiv1.GetMachineInfoReply, error) {
	base, err := normalizeBaseURL(endpoint)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/v1/machine-info", nil)
	if err != nil {
		return nil, err
	}
	c.addAuthHeader(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("get machine info failed, endpoint=%s status=%d body=%s", endpoint, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	out := &apiv1.GetMachineInfoReply{}
	if err := protojson.Unmarshal(body, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) listMachineInfos(ctx context.Context, endpoint string, page, pageSize int32) (*apiv1.GetClusterMachineInfosReply, error) {
	base, err := normalizeBaseURL(endpoint)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/machine-infos?page=%d&pageSize=%d", base, page, pageSize), nil)
	if err != nil {
		return nil, err
	}
	c.addAuthHeader(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("list machine infos failed, endpoint=%s status=%d body=%s", endpoint, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	out := &apiv1.GetClusterMachineInfosReply{}
	if err := protojson.Unmarshal(body, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *apiClient) reportMachineInfos(ctx context.Context, endpoint string, reqBody *apiv1.ReportMachineInfosRequest) error {
	base, err := normalizeBaseURL(endpoint)
	if err != nil {
		return err
	}
	payload, err := protojson.Marshal(reqBody)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/v1/machine-info/report", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.addAuthHeader(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("report machine infos failed, endpoint=%s status=%d body=%s", endpoint, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *apiClient) addAuthHeader(req *http.Request) {
	token := strings.TrimSpace(c.authToken)
	if token == "" {
		return
	}
	if strings.HasPrefix(token, cnst.HTTPHeaderBearerPrefix) {
		req.Header.Set("Authorization", token)
		return
	}
	req.Header.Set("Authorization", strings.Join([]string{cnst.HTTPHeaderBearerPrefix, token}, " "))
}
