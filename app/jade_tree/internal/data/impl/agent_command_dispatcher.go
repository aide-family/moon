package impl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
)

func NewAgentCommandDispatcher() repository.AgentCommandDispatcher {
	return &agentCommandDispatcher{
		client: &http.Client{},
	}
}

type agentCommandDispatcher struct {
	client *http.Client
}

func (d *agentCommandDispatcher) BatchExecute(ctx context.Context, endpoint string, req *bo.BatchExecuteSSHCommandsBo) ([]*bo.BatchExecuteSSHCommandItemBo, error) {
	apiReq := &apiv1.BatchExecuteSSHCommandsRequest{
		Requests: make([]*apiv1.ExecuteSSHCommandRequest, 0, len(req.Requests)),
	}
	for _, item := range req.Requests {
		if item == nil {
			continue
		}
		apiReq.Requests = append(apiReq.Requests, &apiv1.ExecuteSSHCommandRequest{
			CommandUid:     item.CommandUID.Int64(),
			Host:           item.Host,
			Port:           int32(item.Port),
			Username:       item.Username,
			Password:       item.Password,
			PrivateKey:     item.PrivateKey,
			TimeoutSeconds: item.TimeoutSeconds,
		})
	}

	payload, err := protojson.Marshal(apiReq)
	if err != nil {
		return nil, err
	}
	url := normalizeBatchExecuteURL(endpoint)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("agent dispatch status=%d, body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var apiReply apiv1.BatchExecuteSSHCommandsReply
	if err := protojson.Unmarshal(body, &apiReply); err != nil {
		return nil, err
	}
	items := make([]*bo.BatchExecuteSSHCommandItemBo, 0, len(apiReply.GetItems()))
	for _, item := range apiReply.GetItems() {
		if item == nil {
			continue
		}
		boItem := &bo.BatchExecuteSSHCommandItemBo{
			Index: item.GetIndex(),
			Error: item.GetError(),
		}
		if item.GetReply() != nil {
			boItem.Reply = &bo.SSHExecReply{
				Stdout:   item.GetReply().GetStdout(),
				Stderr:   item.GetReply().GetStderr(),
				ExitCode: int(item.GetReply().GetExitCode()),
			}
		}
		items = append(items, boItem)
	}
	return items, nil
}

func normalizeBatchExecuteURL(endpoint string) string {
	endpoint = pickHTTPAddress(endpoint)
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if strings.HasSuffix(endpoint, "/v1/ssh-command-actions/batch-execute") {
		return endpoint
	}
	return endpoint + "/v1/ssh-command-actions/batch-execute"
}

func pickHTTPAddress(endpoint string) string {
	parts := strings.Split(endpoint, ",")
	fallback := ""
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if fallback == "" {
			fallback = part
		}
		if strings.HasPrefix(part, "http://") || strings.HasPrefix(part, "https://") {
			return part
		}
	}
	return fallback
}
