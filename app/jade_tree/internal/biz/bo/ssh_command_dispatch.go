package bo

import (
	"strings"

	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
	"github.com/aide-family/jade_tree/pkg/machine"
)

type BatchExecuteSSHCommandsBo struct {
	Requests []*ExecuteStoredSSHCommandBo
}

type BatchExecuteSSHCommandItemBo struct {
	Index int32
	Reply *SSHExecReply
	Error string
}

type DispatchSSHCommandFilterBo struct {
	IncludeMachineUIDs   []snowflake.ID
	IncludeSystemTypes   []string
	IncludeAgentVersions []string
	ExcludeMachineUIDs   []snowflake.ID
	ExcludeSystemTypes   []string
	ExcludeAgentVersions []string
}

type DispatchSSHCommandToAgentsInput struct {
	CommandUID     snowflake.ID
	Username       string
	Password       string
	PrivateKey     string
	Port           int
	TimeoutSeconds int32
	Filter         *DispatchSSHCommandFilterBo
}

type DispatchSSHCommandResultItemBo struct {
	Machine  *machine.MachineInfo
	Endpoint string
	Reply    *SSHExecReply
	Error    string
}

type DispatchSSHCommandReplyBo struct {
	Total   int64
	Success int64
	Failed  int64
	Items   []*DispatchSSHCommandResultItemBo
}

type CreateProbeTaskDispatchItemBo struct {
	Type           string
	Host           string
	Port           string
	URL            string
	Name           string
	Status         enum.GlobalStatus
	TimeoutSeconds int32
}

type BatchCreateProbeTasksBo struct {
	Requests []*CreateProbeTaskDispatchItemBo
}

type BatchCreateProbeTaskItemResultBo struct {
	Index int32
	UID   snowflake.ID
	Error string
}

type BatchCreateProbeTasksReplyBo struct {
	Items []*BatchCreateProbeTaskItemResultBo
}

func NewDispatchSSHCommandFilterBo(in *apiv1.DispatchSSHCommandFilter) *DispatchSSHCommandFilterBo {
	if in == nil {
		return &DispatchSSHCommandFilterBo{}
	}
	return &DispatchSSHCommandFilterBo{
		IncludeMachineUIDs:   toSnowflakeIDs(in.GetIncludeMachineUids()),
		IncludeSystemTypes:   trimStringList(in.GetIncludeSystemTypes()),
		IncludeAgentVersions: trimStringList(in.GetIncludeAgentVersions()),
		ExcludeMachineUIDs:   toSnowflakeIDs(in.GetExcludeMachineUids()),
		ExcludeSystemTypes:   trimStringList(in.GetExcludeSystemTypes()),
		ExcludeAgentVersions: trimStringList(in.GetExcludeAgentVersions()),
	}
}

func toSnowflakeIDs(ids []int64) []snowflake.ID {
	out := make([]snowflake.ID, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		out = append(out, snowflake.ID(id))
	}
	return out
}

func trimStringList(in []string) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}
