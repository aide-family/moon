package data

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/cluster"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/pkg/merr"
)

var _ repository.ClusterRepo = (*clusterRepo)(nil)

func NewClusterRepo(deps *Dependencies, logger *slog.Logger) repository.ClusterRepo {
	return &clusterRepo{repoBase: newBase(deps, logger), clients: deps.clients}
}

type clusterRepo struct {
	*repoBase
	clients *clusterClients
}

func (r *clusterRepo) CreateCluster(ctx context.Context, in bo.Cluster, kubeconfig string) (*bo.Cluster, error) {
	// 探测不通过就不写库：避免接入一条永远连不上的集群。
	probe, err := probeKubeconfig(ctx, r.log, kubeconfig)
	if err != nil {
		return nil, err
	}
	cipher, err := r.box.encrypt(strings.TrimSpace(kubeconfig))
	if err != nil {
		return nil, err
	}
	row, err := r.client.Cluster.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetServer(probe.Server).
		SetVersion(probe.Version).
		SetClusterName(probe.ClusterName).
		SetContextName(probe.ContextName).
		SetKubeconfigCipher(cipher).
		SetRemark(in.Remark).
		SetLastCheckedAt(time.Now().Unix()).
		SetLastError("").
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toCluster(row)
	return &item, nil
}

func (r *clusterRepo) UpdateCluster(ctx context.Context, in bo.Cluster, kubeconfig string) (*bo.Cluster, error) {
	row, err := r.client.Cluster.Query().Where(
		cluster.NamespaceUIDEQ(in.NamespaceUID),
		cluster.UIDEQ(in.UID),
		cluster.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	update := r.client.Cluster.UpdateOne(row).
		SetName(in.Name).
		SetRemark(in.Remark)
	// 空 kubeconfig 表示保持原凭据和元数据，只改名称与备注。
	replaced := strings.TrimSpace(kubeconfig) != ""
	if replaced {
		probe, err := probeKubeconfig(ctx, r.log, kubeconfig)
		if err != nil {
			return nil, err
		}
		cipher, err := r.box.encrypt(strings.TrimSpace(kubeconfig))
		if err != nil {
			return nil, err
		}
		update.SetKubeconfigCipher(cipher).
			SetServer(probe.Server).
			SetVersion(probe.Version).
			SetClusterName(probe.ClusterName).
			SetContextName(probe.ContextName).
			SetLastCheckedAt(time.Now().Unix()).
			SetLastError("")
	}
	updated, err := update.Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if replaced {
		// 旧 clientset 指向旧凭据，替换后必须丢掉。
		r.clients.invalidate(in.UID)
	}
	item := toCluster(updated)
	return &item, nil
}

func (r *clusterRepo) DeleteCluster(ctx context.Context, namespaceUID, uid string) error {
	n, err := r.client.Cluster.Update().Where(
		cluster.NamespaceUIDEQ(namespaceUID),
		cluster.UIDEQ(uid),
		cluster.DeletedAtIsNil(),
	).SetDeletedAt(time.Now()).SetKubeconfigCipher("").Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	r.clients.invalidate(uid)
	return nil
}

func (r *clusterRepo) GetCluster(ctx context.Context, namespaceUID, uid string) (*bo.Cluster, error) {
	row, err := r.client.Cluster.Query().Where(
		cluster.NamespaceUIDEQ(namespaceUID),
		cluster.UIDEQ(uid),
		cluster.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toCluster(row)
	return &item, nil
}

func (r *clusterRepo) ListClusters(ctx context.Context, namespaceUID, name string, page, pageSize int) ([]bo.Cluster, int, error) {
	query := r.client.Cluster.Query().Where(cluster.NamespaceUIDEQ(namespaceUID), cluster.DeletedAtIsNil())
	if filter := strings.TrimSpace(name); filter != "" {
		query = query.Where(cluster.NameContainsFold(filter))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(cluster.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.Cluster, 0, len(rows))
	for _, row := range rows {
		items = append(items, toCluster(row))
	}
	return items, total, nil
}

func (r *clusterRepo) ProbeKubeconfig(ctx context.Context, kubeconfig string) (*bo.ClusterProbe, error) {
	return probeKubeconfig(ctx, r.log, kubeconfig)
}

func (r *clusterRepo) TestCluster(ctx context.Context, namespaceUID, uid string) (*bo.ClusterProbe, *bo.Cluster, error) {
	row, err := r.client.Cluster.Query().Where(
		cluster.NamespaceUIDEQ(namespaceUID),
		cluster.UIDEQ(uid),
		cluster.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, nil, r.fail(err)
	}
	kubeconfig, err := r.box.decrypt(row.KubeconfigCipher)
	if err != nil {
		r.log.Error(fmt.Sprintf("cluster kubeconfig: %v", err))
		return nil, nil, merr.Server("request failed")
	}
	cipher := row.KubeconfigCipher
	now := time.Now().Unix()
	probe, probeErr := probeKubeconfig(ctx, r.log, kubeconfig)
	if probeErr != nil {
		// 失败也要记录检测时间和原因，页面据此显示集群当前不可用。
		// 只在凭据没有被替换时更新，避免用旧配置的检测结果覆盖新凭据的状态。
		if _, err := r.client.Cluster.Update().Where(
			cluster.NamespaceUIDEQ(namespaceUID),
			cluster.UIDEQ(uid),
			cluster.KubeconfigCipherEQ(cipher),
			cluster.DeletedAtIsNil(),
		).SetLastCheckedAt(now).
			SetLastError(probeMessage(probeErr)).
			Save(ctx); err != nil {
			r.log.Error(fmt.Sprintf("cluster probe state: %v", err))
		}
		return nil, nil, probeErr
	}
	_, err = r.client.Cluster.Update().Where(
		cluster.NamespaceUIDEQ(namespaceUID),
		cluster.UIDEQ(uid),
		cluster.KubeconfigCipherEQ(cipher),
		cluster.DeletedAtIsNil(),
	).SetLastCheckedAt(now).
		SetLastError("").
		SetServer(probe.Server).
		SetVersion(probe.Version).
		SetClusterName(probe.ClusterName).
		SetContextName(probe.ContextName).
		Save(ctx)
	if err != nil {
		return nil, nil, r.fail(err)
	}
	// 探测期间凭据可能已被替换或集群被删除：不覆盖元数据，只按最新记录返回。
	item, err := r.GetCluster(ctx, namespaceUID, uid)
	if err != nil {
		return nil, nil, err
	}
	return probe, item, nil
}

func toCluster(row *ent.Cluster) bo.Cluster {
	return bo.Cluster{
		UID:           row.UID,
		NamespaceUID:  row.NamespaceUID,
		Name:          row.Name,
		Server:        row.Server,
		Version:       row.Version,
		ClusterName:   row.ClusterName,
		ContextName:   row.ContextName,
		Remark:        row.Remark,
		LastCheckedAt: row.LastCheckedAt,
		LastError:     row.LastError,
		CreatedAt:     row.CreatedAt,
	}
}

// probeMessage 取错误文本写进 last_error。它来自 kubeError 的固定文案，
// 不含上游响应或 kubeconfig 内容。
func probeMessage(err error) string {
	return kerrors.FromError(err).Message
}
