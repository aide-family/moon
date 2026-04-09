package biz

import (
	"context"
	"strings"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
	"github.com/aide-family/jade_tree/pkg/machine"
)

type ProbeTask struct {
	repo         repository.ProbeTask
	machineInfos repository.MachineInfoProvider
	dispatcher   repository.AgentCommandDispatcher
	helper       *klog.Helper
}

type pingBatchBuildInput struct {
	SourceID       snowflake.ID
	TimeoutSeconds int32
	TargetIDs      []snowflake.ID
	Targets        map[snowflake.ID]*bo.ProbeTaskItemBo
}

func NewProbeTask(repo repository.ProbeTask, machineInfos repository.MachineInfoProvider, dispatcher repository.AgentCommandDispatcher, helper *klog.Helper) *ProbeTask {
	return &ProbeTask{repo: repo, machineInfos: machineInfos, dispatcher: dispatcher, helper: helper}
}

func (p *ProbeTask) Create(ctx context.Context, in *bo.CreateProbeTaskBo) (*bo.ProbeTaskItemBo, error) {
	if err := p.ensureUnique(ctx, &bo.ProbeTaskUniqueCheckBo{
		Type: in.Fields.Type,
		Host: in.Fields.Host,
		Port: in.Fields.Port,
		URL:  in.Fields.URL,
	}); err != nil {
		return nil, err
	}
	return p.repo.Create(ctx, in)
}

func (p *ProbeTask) Update(ctx context.Context, in *bo.UpdateProbeTaskBo) (*bo.ProbeTaskItemBo, error) {
	if err := p.ensureUnique(ctx, &bo.ProbeTaskUniqueCheckBo{
		Type:       in.Fields.Type,
		Host:       in.Fields.Host,
		Port:       in.Fields.Port,
		URL:        in.Fields.URL,
		ExcludeUID: in.UID,
	}); err != nil {
		return nil, err
	}
	return p.repo.Update(ctx, in)
}

func (p *ProbeTask) UpdateStatus(ctx context.Context, in *bo.UpdateProbeTaskStatusBo) (*bo.ProbeTaskItemBo, error) {
	return p.repo.UpdateStatus(ctx, in)
}

func (p *ProbeTask) Delete(ctx context.Context, uid snowflake.ID) error {
	return p.repo.Delete(ctx, uid)
}

func (p *ProbeTask) Get(ctx context.Context, uid snowflake.ID) (*bo.ProbeTaskItemBo, error) {
	return p.repo.Get(ctx, uid)
}

func (p *ProbeTask) List(ctx context.Context, req *bo.ListProbeTasksBo) (*bo.PageResponseBo[*bo.ProbeTaskItemBo], error) {
	return p.repo.List(ctx, req)
}

func (p *ProbeTask) CreatePingProbeTasks(ctx context.Context, in *bo.CreatePingProbeTasksInput) (*bo.DispatchCreateProbeTasksReplyBo, error) {
	if in == nil {
		return nil, merr.ErrorInvalidArgument("ping input is required")
	}
	timeoutSeconds := in.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 5
	}
	local, localErr := p.machineInfos.GetMachineInfoByIdentity(ctx, p.machineInfos.GetLocalMachineIdentity())
	if localErr != nil && !merr.IsNotFound(localErr) {
		return nil, localErr
	}
	localID := snowflake.ID(0)
	if local != nil {
		localID = local.ID
	}
	sourceMap := make(map[snowflake.ID]struct{}, len(in.SourceMachineUIDs))
	targetMap := make(map[snowflake.ID]struct{}, len(in.TargetMachineUIDs))
	var prefetchedMachines []*machine.MachineInfo
	for _, id := range in.SourceMachineUIDs {
		if id > 0 {
			sourceMap[id] = struct{}{}
		}
	}
	for _, id := range in.TargetMachineUIDs {
		if id > 0 {
			targetMap[id] = struct{}{}
		}
	}
	switch {
	case len(sourceMap) > 0 && len(targetMap) == 0:
		for id := range sourceMap {
			targetMap[id] = struct{}{}
		}
	case len(sourceMap) == 0 && len(targetMap) > 0:
		if localID > 0 {
			sourceMap[localID] = struct{}{}
		}
	case len(sourceMap) == 0 && len(targetMap) == 0:
		if localID > 0 {
			sourceMap[localID] = struct{}{}
		}
		allMachines, err := p.machineInfos.ListDispatchTargets(ctx, nil)
		if err != nil {
			return nil, err
		}
		prefetchedMachines = allMachines
		for _, m := range allMachines {
			if m == nil || m.ID <= 0 {
				continue
			}
			targetMap[m.ID] = struct{}{}
		}
	}
	if len(sourceMap) == 0 {
		return nil, merr.ErrorInvalidArgument("source machines are empty")
	}
	if len(targetMap) == 0 {
		return nil, merr.ErrorInvalidArgument("target machines are empty")
	}

	if prefetchedMachines == nil {
		queryIDs := make(map[snowflake.ID]struct{}, len(sourceMap)+len(targetMap))
		for id := range sourceMap {
			queryIDs[id] = struct{}{}
		}
		for id := range targetMap {
			queryIDs[id] = struct{}{}
		}
		allNeededMachines, err := p.machineInfos.ListDispatchTargets(ctx, &bo.DispatchSSHCommandFilterBo{
			IncludeMachineUIDs: mapKeys(queryIDs),
		})
		if err != nil {
			return nil, err
		}
		prefetchedMachines = allNeededMachines
	}

	machineByID := make(map[snowflake.ID]*machine.MachineInfo, len(prefetchedMachines))
	for _, m := range prefetchedMachines {
		if m == nil || m.ID <= 0 {
			continue
		}
		machineByID[m.ID] = m
	}

	targetByID := make(map[snowflake.ID]*bo.DispatchCreateProbeTaskResultItemBo, len(targetMap))
	targetMachine := make(map[snowflake.ID]*bo.ProbeTaskItemBo)
	for id := range targetMap {
		m, ok := machineByID[id]
		if !ok {
			continue
		}
		ip := ""
		if m.Network != nil {
			ip = strings.TrimSpace(m.Network.LocalIP)
		}
		targetByID[m.ID] = &bo.DispatchCreateProbeTaskResultItemBo{
			MachineUID:  m.ID,
			MachineUUID: m.MachineUUID,
			HostName:    m.HostName,
			LocalIP:     ip,
		}
		targetMachine[id] = &bo.ProbeTaskItemBo{UID: id, Host: ip, Name: m.HostName}
	}
	if len(targetByID) == 0 {
		return nil, merr.ErrorInvalidArgument("no target machines found")
	}
	out := &bo.DispatchCreateProbeTasksReplyBo{
		Total: int64(len(sourceMap)),
		Items: make([]*bo.DispatchCreateProbeTaskResultItemBo, 0, len(sourceMap)),
	}
	localTargets := make([]snowflake.ID, 0, len(targetByID))
	for id := range targetByID {
		localTargets = append(localTargets, id)
	}
	for sourceID := range sourceMap {
		item := &bo.DispatchCreateProbeTaskResultItemBo{MachineUID: sourceID}
		batchReq := p.buildPingTaskBatch(&pingBatchBuildInput{
			SourceID:       sourceID,
			TimeoutSeconds: timeoutSeconds,
			TargetIDs:      localTargets,
			Targets:        targetMachine,
		})
		if len(batchReq.Requests) == 0 {
			item.Error = "no target machines available for ping task creation"
			out.Failed++
			out.Items = append(out.Items, item)
			continue
		}
		if sourceID == localID && sourceID > 0 {
			item.HostName = local.HostName
			if local.Network != nil {
				item.LocalIP = local.Network.LocalIP
			}
			item.MachineUUID = local.MachineUUID
			count, createErr := p.createPingTasksForLocalSource(ctx, batchReq)
			item.CreatedCount = count
			if createErr != nil {
				item.Error = createErr.Error()
				out.Failed++
			} else {
				out.Success++
				out.CreatedCount += count
			}
			out.Items = append(out.Items, item)
			continue
		}
		sourceMachine, ok := machineByID[sourceID]
		if !ok {
			item.Error = "source machine not found"
			out.Failed++
			out.Items = append(out.Items, item)
			continue
		}
		item.MachineUUID = sourceMachine.MachineUUID
		item.HostName = sourceMachine.HostName
		if sourceMachine.Network != nil {
			item.LocalIP = sourceMachine.Network.LocalIP
		}
		if sourceMachine.Agent == nil || strings.TrimSpace(selectAgentEndpoint(sourceMachine.Agent)) == "" {
			item.Error = "agent endpoint is required"
			out.Failed++
			out.Items = append(out.Items, item)
			continue
		}
		item.Endpoint = selectAgentEndpoint(sourceMachine.Agent)
		reply, dispatchErr := p.dispatcher.BatchCreateProbeTasks(ctx, sourceMachine.Agent, batchReq)
		if dispatchErr != nil {
			item.Error = dispatchErr.Error()
			out.Failed++
			out.Items = append(out.Items, item)
			continue
		}
		createdCount, rowErr := summarizeBatchCreateProbeTasks(reply)
		item.CreatedCount = createdCount
		if rowErr != "" {
			item.Error = rowErr
			out.Failed++
		} else {
			out.Success++
			out.CreatedCount += createdCount
		}
		out.Items = append(out.Items, item)
	}
	out.Total = int64(len(out.Items))
	return out, nil
}

func (p *ProbeTask) ensureUnique(ctx context.Context, in *bo.ProbeTaskUniqueCheckBo) error {
	if in == nil {
		return merr.ErrorInvalidArgument("probe task unique check input is required")
	}
	count, err := p.repo.CountByTypeTarget(ctx, in)
	if err != nil {
		return err
	}
	if count > 0 {
		return merr.ErrorInvalidArgument("probe task already exists for same type and target")
	}
	return nil
}

func mapKeys(in map[snowflake.ID]struct{}) []snowflake.ID {
	out := make([]snowflake.ID, 0, len(in))
	for id := range in {
		out = append(out, id)
	}
	return out
}

func (p *ProbeTask) buildPingTaskBatch(in *pingBatchBuildInput) *bo.BatchCreateProbeTasksBo {
	if in == nil {
		return &bo.BatchCreateProbeTasksBo{}
	}
	req := &bo.BatchCreateProbeTasksBo{Requests: make([]*bo.CreateProbeTaskDispatchItemBo, 0, len(in.TargetIDs))}
	for _, targetID := range in.TargetIDs {
		if targetID == in.SourceID {
			continue
		}
		target, ok := in.Targets[targetID]
		if !ok || strings.TrimSpace(target.Host) == "" {
			continue
		}
		req.Requests = append(req.Requests, &bo.CreateProbeTaskDispatchItemBo{
			Type:           "ping",
			Host:           target.Host,
			Name:           target.Name,
			Status:         enum.GlobalStatus_ENABLED,
			TimeoutSeconds: in.TimeoutSeconds,
		})
	}
	return req
}

func summarizeBatchCreateProbeTasks(in *bo.BatchCreateProbeTasksReplyBo) (int64, string) {
	if in == nil {
		return 0, "empty dispatch response"
	}
	var created int64
	var firstErr string
	for _, item := range in.Items {
		if item == nil {
			continue
		}
		if item.Error != "" && firstErr == "" {
			firstErr = item.Error
			continue
		}
		if item.UID > 0 {
			created++
		}
	}
	return created, firstErr
}

func (p *ProbeTask) createPingTasksForLocalSource(ctx context.Context, batch *bo.BatchCreateProbeTasksBo) (int64, error) {
	if batch == nil {
		return 0, merr.ErrorInvalidArgument("ping batch is required")
	}
	var created int64
	var firstErr error
	for _, item := range batch.Requests {
		createIn := &bo.CreateProbeTaskBo{
			Fields: bo.ProbeTaskFieldsBo{
				Type:           item.Type,
				Host:           item.Host,
				Name:           item.Name,
				Status:         item.Status,
				TimeoutSeconds: item.TimeoutSeconds,
			},
		}
		if err := p.ensureUnique(ctx, &bo.ProbeTaskUniqueCheckBo{
			Type: createIn.Fields.Type,
			Host: createIn.Fields.Host,
		}); err != nil {
			if firstErr == nil && !merr.IsInvalidArgument(err) {
				firstErr = err
			}
			continue
		}
		if _, err := p.repo.Create(ctx, createIn); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		created++
	}
	if created == 0 && firstErr != nil {
		return 0, firstErr
	}
	return created, nil
}
