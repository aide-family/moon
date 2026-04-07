package bo

import (
	"strings"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
)

type ProbeTaskItemBo struct {
	UID            snowflake.ID
	Type           string
	Host           string
	Port           string
	URL            string
	Name           string
	Status         enum.GlobalStatus
	TimeoutSeconds int32
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ProbeTaskFieldsBo struct {
	Type           string
	Host           string
	Port           string
	URL            string
	Name           string
	Status         enum.GlobalStatus
	TimeoutSeconds int32
}

type CreateProbeTaskBo struct {
	Creator snowflake.ID
	Fields  ProbeTaskFieldsBo
}

type UpdateProbeTaskBo struct {
	UID    snowflake.ID
	Fields ProbeTaskUpdateFieldsBo
}

type ProbeTaskUpdateFieldsBo struct {
	Type           string
	Host           string
	Port           string
	URL            string
	Name           string
	TimeoutSeconds int32
}

type UpdateProbeTaskStatusBo struct {
	UID    snowflake.ID
	Status enum.GlobalStatus
}

type ListProbeTasksBo struct {
	*PageRequestBo
	Type    string
	Keyword string
	Status  enum.GlobalStatus
}

func NewCreateProbeTaskBo(req *apiv1.CreateProbeTaskRequest, creator snowflake.ID) (*CreateProbeTaskBo, error) {
	fields, err := validateProbeTaskFields(&ProbeTaskFieldsBo{
		Type:           req.GetType(),
		Host:           req.GetHost(),
		Port:           req.GetPort(),
		URL:            req.GetUrl(),
		Name:           req.GetName(),
		Status:         req.GetStatus(),
		TimeoutSeconds: req.GetTimeoutSeconds(),
	})
	if err != nil {
		return nil, err
	}
	return &CreateProbeTaskBo{Creator: creator, Fields: *fields}, nil
}

func NewUpdateProbeTaskBo(req *apiv1.UpdateProbeTaskRequest) (*UpdateProbeTaskBo, error) {
	fields, err := validateProbeTaskUpdateFields(&ProbeTaskUpdateFieldsBo{
		Type:           req.GetType(),
		Host:           req.GetHost(),
		Port:           req.GetPort(),
		URL:            req.GetUrl(),
		Name:           req.GetName(),
		TimeoutSeconds: req.GetTimeoutSeconds(),
	})
	if err != nil {
		return nil, err
	}
	return &UpdateProbeTaskBo{UID: snowflake.ID(req.GetUid()), Fields: *fields}, nil
}

func NewListProbeTasksBo(req *apiv1.ListProbeTasksRequest) *ListProbeTasksBo {
	return &ListProbeTasksBo{
		PageRequestBo: NewPageRequestBo(req.GetPage(), req.GetPageSize()),
		Type:          strings.ToLower(strings.TrimSpace(req.GetType())),
		Keyword:       strings.TrimSpace(req.GetKeyword()),
		Status:        req.GetStatus(),
	}
}

func NewUpdateProbeTaskStatusBo(req *apiv1.UpdateProbeTaskStatusRequest) (*UpdateProbeTaskStatusBo, error) {
	status := req.GetStatus()
	if status != enum.GlobalStatus_ENABLED && status != enum.GlobalStatus_DISABLED {
		return nil, merr.ErrorInvalidArgument("status must be ENABLED or DISABLED")
	}
	return &UpdateProbeTaskStatusBo{
		UID:    snowflake.ID(req.GetUid()),
		Status: status,
	}, nil
}

func ToAPIV1ProbeTaskItem(in *ProbeTaskItemBo) *apiv1.ProbeTaskItem {
	if in == nil {
		return nil
	}
	return &apiv1.ProbeTaskItem{
		Uid:            in.UID.Int64(),
		Type:           in.Type,
		Host:           in.Host,
		Port:           in.Port,
		Url:            in.URL,
		Name:           in.Name,
		Status:         in.Status,
		TimeoutSeconds: in.TimeoutSeconds,
		CreatedAt:      timex.FormatTime(&in.CreatedAt),
		UpdatedAt:      timex.FormatTime(&in.UpdatedAt),
	}
}

func validateProbeTaskFields(in *ProbeTaskFieldsBo) (*ProbeTaskFieldsBo, error) {
	if in == nil {
		return nil, merr.ErrorInvalidArgument("probe task fields are required")
	}
	in.Type = strings.ToLower(strings.TrimSpace(in.Type))
	in.Host = strings.TrimSpace(in.Host)
	in.Port = strings.TrimSpace(in.Port)
	in.URL = strings.TrimSpace(in.URL)
	in.Name = strings.TrimSpace(in.Name)
	if in.Status == enum.GlobalStatus_GlobalStatus_UNKNOWN {
		in.Status = enum.GlobalStatus_ENABLED
	}
	if in.Status != enum.GlobalStatus_ENABLED && in.Status != enum.GlobalStatus_DISABLED {
		return nil, merr.ErrorInvalidArgument("status must be ENABLED or DISABLED")
	}
	if in.TimeoutSeconds <= 0 {
		in.TimeoutSeconds = 5
	}
	switch in.Type {
	case "tcp", "port":
		if in.Host == "" || in.Port == "" {
			return nil, merr.ErrorInvalidArgument("host and port are required")
		}
	case "http":
		if in.URL == "" {
			return nil, merr.ErrorInvalidArgument("url is required")
		}
	case "cert":
		if in.Host == "" {
			return nil, merr.ErrorInvalidArgument("host is required")
		}
		if in.Port == "" {
			in.Port = "443"
		}
	default:
		return nil, merr.ErrorInvalidArgument("type must be tcp, port, http or cert")
	}
	if in.Name == "" {
		if in.Type == "http" {
			in.Name = in.URL
		} else {
			in.Name = in.Host + ":" + in.Port
		}
	}
	return in, nil
}

func validateProbeTaskUpdateFields(in *ProbeTaskUpdateFieldsBo) (*ProbeTaskUpdateFieldsBo, error) {
	if in == nil {
		return nil, merr.ErrorInvalidArgument("probe task update fields are required")
	}
	in.Type = strings.ToLower(strings.TrimSpace(in.Type))
	in.Host = strings.TrimSpace(in.Host)
	in.Port = strings.TrimSpace(in.Port)
	in.URL = strings.TrimSpace(in.URL)
	in.Name = strings.TrimSpace(in.Name)
	if in.TimeoutSeconds <= 0 {
		in.TimeoutSeconds = 5
	}
	switch in.Type {
	case "tcp", "port":
		if in.Host == "" || in.Port == "" {
			return nil, merr.ErrorInvalidArgument("host and port are required")
		}
	case "http":
		if in.URL == "" {
			return nil, merr.ErrorInvalidArgument("url is required")
		}
	case "cert":
		if in.Host == "" {
			return nil, merr.ErrorInvalidArgument("host is required")
		}
		if in.Port == "" {
			in.Port = "443"
		}
	default:
		return nil, merr.ErrorInvalidArgument("type must be tcp, port, http or cert")
	}
	if in.Name == "" {
		if in.Type == "http" {
			in.Name = in.URL
		} else {
			in.Name = in.Host + ":" + in.Port
		}
	}
	return in, nil
}
