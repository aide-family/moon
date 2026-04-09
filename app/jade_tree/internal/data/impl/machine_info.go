// Package impl provides repository implementations.
package impl

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"
	"strings"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/jade_tree/cmd"
	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
	"github.com/aide-family/jade_tree/internal/conf"
	"github.com/aide-family/jade_tree/internal/data"
	"github.com/aide-family/jade_tree/internal/data/impl/query"
	"github.com/aide-family/jade_tree/pkg/machine"
)

func NewMachineInfoRepository(bc *conf.Bootstrap, d *data.Data) repository.MachineInfoProvider {
	query.SetDefault(d.DB())
	var enabledCollectSelf bool
	if bc != nil && bc.GetCollectSelf() != nil {
		enabledCollectSelf = strings.EqualFold(bc.GetCollectSelf().GetEnabled(), "true")
	}
	return &machineInfoRepository{Data: d, enabledCollectSelf: enabledCollectSelf}
}

type machineInfoRepository struct {
	*data.Data
	enabledCollectSelf bool
}

func (m *machineInfoRepository) Collect(ctx context.Context) (*machine.MachineInfo, error) {
	if !m.enabledCollectSelf {
		return nil, merr.ErrorParams("collect self is not enabled")
	}
	return machine.Collect(ctx)
}

func (m *machineInfoRepository) GetLocalMachineIdentity() *bo.MachineInfoIdentityBo {
	u, h, lip := machine.LocalMachineIdentity()
	return &bo.MachineInfoIdentityBo{MachineUUID: u, HostName: h, LocalIP: lip}
}

func (m *machineInfoRepository) GetLocalAgentInfo(localIP string) *machine.MachineAgent {
	version := strings.TrimSpace(cmd.GetGlobalFlags().Version)
	if version == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			version = strings.TrimSpace(info.Main.Version)
		}
	}
	if version == "" {
		version = "latest"
	}

	httpEndpoint, grpcEndpoint := cmd.GetServerEndpoints()
	if httpEndpoint == "" {
		httpEndpoint = strings.TrimSpace(m.Config().GetServer().GetHttp().GetAddress())
	}
	if grpcEndpoint == "" {
		grpcEndpoint = strings.TrimSpace(m.Config().GetServer().GetGrpc().GetAddress())
	}
	httpEndpoint = m.normalizeServerEndpoint("http", httpEndpoint, localIP)
	grpcEndpoint = m.normalizeServerEndpoint("grpc", grpcEndpoint, localIP)
	merged := make([]string, 0, 2)
	if httpEndpoint != "" {
		merged = append(merged, httpEndpoint)
	}
	if grpcEndpoint != "" {
		merged = append(merged, grpcEndpoint)
	}
	if len(merged) == 0 {
		return &machine.MachineAgent{Version: version}
	}

	return &machine.MachineAgent{
		Endpoint:     strings.Join(merged, ","),
		HTTPEndpoint: httpEndpoint,
		GRPCEndpoint: grpcEndpoint,
		Version:      version,
	}
}

func (m *machineInfoRepository) normalizeServerEndpoint(scheme, address, localIP string) string {
	address = strings.TrimSpace(address)
	if address == "" {
		return ""
	}
	if strings.Contains(address, "://") {
		return address
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Sprintf("%s://%s", scheme, address)
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = strings.TrimSpace(localIP)
		if host == "" {
			host = "127.0.0.1"
		}
	}
	return fmt.Sprintf("%s://%s:%s", scheme, host, port)
}
