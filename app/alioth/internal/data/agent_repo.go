package data

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/agent"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

func (r *agentRepo) CreateAgent(ctx context.Context, in bo.Agent, tokenHash string) (*bo.Agent, error) {
	password, err := r.box.encrypt(in.SSHPassword)
	if err != nil {
		return nil, err
	}
	key, err := r.box.encrypt(in.SSHPrivateKey)
	if err != nil {
		return nil, err
	}
	port := in.SSHPort
	if port <= 0 {
		port = 22
	}
	row, err := r.client.Agent.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetStatus(agent.StatusACTIVE).
		SetTokenHash(tokenHash).
		SetSSHHost(in.SSHHost).
		SetSSHPort(port).
		SetSSHUser(in.SSHUser).
		SetSSHPasswordCipher(password).
		SetSSHPrivateKeyCipher(key).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toAgent(row)
	return &item, nil
}

func (r *agentRepo) UpdateAgent(ctx context.Context, in bo.Agent) (*bo.Agent, error) {
	row, err := r.client.Agent.Query().Where(agent.NamespaceUIDEQ(in.NamespaceUID), agent.UIDEQ(in.UID), agent.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	password := row.SSHPasswordCipher
	if in.SSHPassword != "" {
		password, err = r.box.encrypt(in.SSHPassword)
		if err != nil {
			return nil, err
		}
	}
	key := row.SSHPrivateKeyCipher
	if in.SSHPrivateKey != "" {
		key, err = r.box.encrypt(in.SSHPrivateKey)
		if err != nil {
			return nil, err
		}
	}
	port := in.SSHPort
	if port <= 0 {
		port = 22
	}
	updated, err := r.client.Agent.UpdateOne(row).
		SetName(in.Name).
		SetSSHHost(in.SSHHost).
		SetSSHPort(port).
		SetSSHUser(in.SSHUser).
		SetSSHPasswordCipher(password).
		SetSSHPrivateKeyCipher(key).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toAgent(updated)
	return &item, nil
}

func (r *agentRepo) SetTokenHash(ctx context.Context, namespaceUID, uid, tokenHash, status string) (*bo.Agent, error) {
	value := agent.StatusACTIVE
	if status == "REVOKED" {
		value = agent.StatusREVOKED
	}
	n, err := r.client.Agent.Update().Where(agent.NamespaceUIDEQ(namespaceUID), agent.UIDEQ(uid), agent.DeletedAtIsNil()).
		SetTokenHash(tokenHash).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetAgent(ctx, namespaceUID, uid)
}

func (r *agentRepo) DeleteAgent(ctx context.Context, namespaceUID, uid string) error {
	n, err := r.client.Agent.Update().Where(agent.NamespaceUIDEQ(namespaceUID), agent.UIDEQ(uid), agent.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).SetOnline(false).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	return nil
}

func (r *agentRepo) GetAgent(ctx context.Context, namespaceUID, uid string) (*bo.Agent, error) {
	row, err := r.client.Agent.Query().Where(agent.NamespaceUIDEQ(namespaceUID), agent.UIDEQ(uid), agent.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toAgent(row)
	return &item, nil
}

func (r *agentRepo) GetAgentByUID(ctx context.Context, uid string) (*bo.Agent, error) {
	row, err := r.client.Agent.Query().Where(agent.UIDEQ(uid), agent.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toAgent(row)
	return &item, nil
}

func (r *agentRepo) GetAgentByTokenHash(ctx context.Context, tokenHash string) (*bo.Agent, error) {
	row, err := r.client.Agent.Query().Where(agent.TokenHashEQ(tokenHash), agent.DeletedAtIsNil(), agent.StatusEQ(agent.StatusACTIVE)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toAgent(row)
	return &item, nil
}

func (r *agentRepo) ListAgents(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.Agent, int, error) {
	query := r.client.Agent.Query().Where(agent.NamespaceUIDEQ(namespaceUID), agent.DeletedAtIsNil())
	if name != "" {
		query = query.Where(agent.Or(agent.NameContainsFold(name), agent.HostnameContainsFold(name)))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(agent.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.Agent, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAgent(row))
	}
	return items, total, nil
}

func (r *agentRepo) SetOnline(ctx context.Context, uid string, online bool, hostname, machineID, localIP, version, collectJSON string) error {
	upd := r.client.Agent.Update().Where(agent.UIDEQ(uid), agent.DeletedAtIsNil()).SetOnline(online)
	if hostname != "" {
		upd = upd.SetHostname(hostname)
	}
	if machineID != "" {
		upd = upd.SetMachineID(machineID)
	}
	if localIP != "" {
		upd = upd.SetLocalIP(localIP)
	}
	if version != "" {
		upd = upd.SetVersion(version)
	}
	if collectJSON != "" {
		upd = upd.SetCollectJSON(collectJSON)
	}
	_, err := upd.Save(ctx)
	return r.fail(err)
}

func (r *agentRepo) GetSSHSecrets(ctx context.Context, uid string) (string, string, error) {
	row, err := r.client.Agent.Query().Where(agent.UIDEQ(uid), agent.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return "", "", r.fail(err)
	}
	password, err := r.box.decrypt(row.SSHPasswordCipher)
	if err != nil {
		return "", "", err
	}
	key, err := r.box.decrypt(row.SSHPrivateKeyCipher)
	if err != nil {
		return "", "", err
	}
	return password, key, nil
}

func toAgent(row *ent.Agent) bo.Agent {
	return bo.Agent{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name, Hostname: row.Hostname,
		MachineID: row.MachineID, LocalIP: row.LocalIP, Version: row.Version, CollectJSON: row.CollectJSON,
		Online: row.Online, Status: string(row.Status), SSHHost: row.SSHHost, SSHPort: row.SSHPort,
		SSHUser: row.SSHUser, HasPassword: row.SSHPasswordCipher != "", HasPrivateKey: row.SSHPrivateKeyCipher != "",
		Remark: row.Remark,
	}
}
