package data

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/registry"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

var _ repository.RegistryRepo = (*registryRepo)(nil)

func NewRegistryRepo(deps *Dependencies, logger *slog.Logger, images repository.ImageRegistryRepo) repository.RegistryRepo {
	return &registryRepo{repoBase: newBase(deps, logger), images: images}
}

type registryRepo struct {
	*repoBase
	images repository.ImageRegistryRepo
}

func (r *registryRepo) CreateRegistry(ctx context.Context, in bo.Registry, password string) (*bo.Registry, error) {
	// 先探测再入库，避免保存一条永远连不上的连接。
	probe, err := r.images.Probe(ctx, in.Driver, in.Endpoint, in.Username, password, in.AllowInsecureTLS)
	if err != nil {
		return nil, err
	}
	cipher, err := r.box.encrypt(password)
	if err != nil {
		return nil, err
	}
	row, err := r.client.Registry.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetDriver(registry.Driver(in.Driver)).
		SetEndpoint(strings.TrimSuffix(strings.TrimSpace(in.Endpoint), "/")).
		SetUsername(in.Username).
		SetPasswordCipher(cipher).
		SetAllowInsecureTLS(in.AllowInsecureTLS).
		SetRemark(in.Remark).
		SetLastCheckedAt(time.Now().Unix()).
		SetLastError("").
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	_ = probe
	item := toRegistry(row)
	return &item, nil
}

func (r *registryRepo) UpdateRegistry(ctx context.Context, in bo.Registry, password string, clearPassword bool) (*bo.Registry, error) {
	row, err := r.client.Registry.Query().Where(
		registry.UIDEQ(in.UID), registry.NamespaceUIDEQ(in.NamespaceUID), registry.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	effectivePassword := password
	if !clearPassword && strings.TrimSpace(password) == "" {
		effectivePassword, err = r.box.decrypt(row.PasswordCipher)
		if err != nil {
			return nil, merr.Server("request failed")
		}
	}
	if _, err := r.images.Probe(ctx, in.Driver, in.Endpoint, in.Username, effectivePassword, in.AllowInsecureTLS); err != nil {
		return nil, err
	}
	update := r.client.Registry.UpdateOne(row).
		SetName(in.Name).
		SetDriver(registry.Driver(in.Driver)).
		SetEndpoint(strings.TrimSuffix(strings.TrimSpace(in.Endpoint), "/")).
		SetUsername(in.Username).
		SetAllowInsecureTLS(in.AllowInsecureTLS).
		SetRemark(in.Remark).
		SetLastCheckedAt(time.Now().Unix()).
		SetLastError("")
	if clearPassword {
		update.SetPasswordCipher("")
	} else if strings.TrimSpace(password) != "" {
		cipher, err := r.box.encrypt(password)
		if err != nil {
			return nil, err
		}
		update.SetPasswordCipher(cipher)
	}
	updated, err := update.Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toRegistry(updated)
	return &item, nil
}

func (r *registryRepo) DeleteRegistry(ctx context.Context, namespaceUID, uid string) error {
	now := time.Now().UTC()
	n, err := r.client.Registry.Update().
		Where(registry.UIDEQ(uid), registry.NamespaceUIDEQ(namespaceUID), registry.DeletedAtIsNil()).
		SetDeletedAt(now).
		// 删除后不保留可用于连接的凭据。
		SetPasswordCipher("").
		Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("registry not found")
	}
	return nil
}

func (r *registryRepo) GetRegistry(ctx context.Context, namespaceUID, uid string) (*bo.Registry, error) {
	row, err := r.client.Registry.Query().Where(
		registry.UIDEQ(uid), registry.NamespaceUIDEQ(namespaceUID), registry.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toRegistry(row)
	return &item, nil
}

func (r *registryRepo) ListRegistries(ctx context.Context, namespaceUID string, filter bo.RegistryListFilter, page, pageSize int) ([]bo.Registry, int, error) {
	query := r.client.Registry.Query().Where(
		registry.NamespaceUIDEQ(namespaceUID), registry.DeletedAtIsNil(),
	)
	if name := strings.TrimSpace(filter.Name); name != "" {
		query.Where(registry.NameContainsFold(name))
	}
	if driver := strings.TrimSpace(filter.Driver); driver != "" {
		query.Where(registry.DriverEQ(registry.Driver(driver)))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(registry.FieldCreatedAt)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.Registry, 0, len(rows))
	for _, row := range rows {
		items = append(items, toRegistry(row))
	}
	return items, total, nil
}

func (r *registryRepo) TouchRegistry(ctx context.Context, namespaceUID, uid string, checkedAt int64, lastError string) (*bo.Registry, error) {
	_, err := r.client.Registry.Update().
		Where(registry.UIDEQ(uid), registry.NamespaceUIDEQ(namespaceUID), registry.DeletedAtIsNil()).
		SetLastCheckedAt(checkedAt).
		SetLastError(lastError).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.GetRegistry(ctx, namespaceUID, uid)
}

func toRegistry(row *ent.Registry) bo.Registry {
	return bo.Registry{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name,
		Driver: string(row.Driver), Endpoint: row.Endpoint, Username: row.Username,
		AllowInsecureTLS: row.AllowInsecureTLS, Remark: row.Remark,
		LastCheckedAt: row.LastCheckedAt, LastError: row.LastError, CreatedAt: row.CreatedAt,
	}
}

func (r *registryRepo) registryPassword(ctx context.Context, uid string) (string, error) {
	row, err := r.client.Registry.Query().Where(registry.UIDEQ(uid), registry.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return "", r.fail(err)
	}
	plain, err := r.box.decrypt(row.PasswordCipher)
	if err != nil {
		r.log.Error(fmt.Sprintf("registry credential: %v", err))
		return "", merr.Server("request failed")
	}
	return plain, nil
}
