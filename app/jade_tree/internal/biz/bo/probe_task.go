package bo

import (
	"strings"
	"time"

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
	Enabled        bool
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
	Enabled        bool
	TimeoutSeconds int32
}

type CreateProbeTaskBo struct {
	Creator snowflake.ID
	Fields  ProbeTaskFieldsBo
}

type UpdateProbeTaskBo struct {
	UID    snowflake.ID
	Fields ProbeTaskFieldsBo
}

type ListProbeTasksBo struct {
	*PageRequestBo
}

func NewCreateProbeTaskBo(req *apiv1.CreateProbeTaskRequest, creator snowflake.ID) (*CreateProbeTaskBo, error) {
	fields, err := validateProbeTaskFields(&ProbeTaskFieldsBo{
		Type:           req.GetType(),
		Host:           req.GetHost(),
		Port:           req.GetPort(),
		URL:            req.GetUrl(),
		Name:           req.GetName(),
		Enabled:        req.GetEnabled(),
		TimeoutSeconds: req.GetTimeoutSeconds(),
	})
	if err != nil {
		return nil, err
	}
	return &CreateProbeTaskBo{Creator: creator, Fields: *fields}, nil
}

func NewUpdateProbeTaskBo(req *apiv1.UpdateProbeTaskRequest) (*UpdateProbeTaskBo, error) {
	fields, err := validateProbeTaskFields(&ProbeTaskFieldsBo{
		Type:           req.GetType(),
		Host:           req.GetHost(),
		Port:           req.GetPort(),
		URL:            req.GetUrl(),
		Name:           req.GetName(),
		Enabled:        req.GetEnabled(),
		TimeoutSeconds: req.GetTimeoutSeconds(),
	})
	if err != nil {
		return nil, err
	}
	return &UpdateProbeTaskBo{UID: snowflake.ID(req.GetUid()), Fields: *fields}, nil
}

func NewListProbeTasksBo(page, pageSize int32) *ListProbeTasksBo {
	return &ListProbeTasksBo{PageRequestBo: NewPageRequestBo(page, pageSize)}
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
		Enabled:        in.Enabled,
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
