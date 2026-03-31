package biz

import (
	"context"
	"strconv"
	"slices"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"
	rabbitv1 "github.com/aide-family/rabbit/pkg/api/v1"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

type ttlCacheItem struct {
	value    string
	expires  int64 // unix nano
}

// notificationResourceNameTTLCache caches remote resource names by uid for a short TTL.
// Key format: "<resourceType>:<uid>".
type notificationResourceNameTTLCache struct {
	mu   sync.RWMutex
	data map[string]ttlCacheItem
}

func newNotificationResourceNameTTLCache() *notificationResourceNameTTLCache {
	return &notificationResourceNameTTLCache{
		data: make(map[string]ttlCacheItem),
	}
}

func (c *notificationResourceNameTTLCache) get(key string, nowUnixNano int64) (string, bool) {
	c.mu.RLock()
	item, ok := c.data[key]
	c.mu.RUnlock()
	if !ok {
		return "", false
	}
	if item.expires > 0 && item.expires <= nowUnixNano {
		// Lazy expire
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		return "", false
	}
	return item.value, true
}

func (c *notificationResourceNameTTLCache) set(key, value string, ttl time.Duration, nowUnixNano int64) {
	expires := nowUnixNano + ttl.Nanoseconds()
	c.mu.Lock()
	c.data[key] = ttlCacheItem{
		value:   value,
		expires: expires,
	}
	c.mu.Unlock()
}

var notificationResourceNameTTLCacheStore = newNotificationResourceNameTTLCache()

func NewNotificationGroup(
	repo repository.NotificationGroup,
	memberRepo repository.Member,
	rabbitWebhookRepo repository.RabbitWebhook,
	rabbitTemplateRepo repository.RabbitTemplate,
	rabbitEmailRepo repository.RabbitEmail,
	helper *klog.Helper,
) *NotificationGroupBiz {
	return &NotificationGroupBiz{
		repo:               repo,
		memberRepo:         memberRepo,
		rabbitWebhookRepo:  rabbitWebhookRepo,
		rabbitTemplateRepo: rabbitTemplateRepo,
		rabbitEmailRepo:    rabbitEmailRepo,
		helper:             klog.NewHelper(klog.With(helper.Logger(), "biz", "notification_group")),
	}
}

type NotificationGroupBiz struct {
	repo               repository.NotificationGroup
	memberRepo         repository.Member
	rabbitWebhookRepo  repository.RabbitWebhook
	rabbitTemplateRepo repository.RabbitTemplate
	rabbitEmailRepo    repository.RabbitEmail
	helper             *klog.Helper
}

func (b *NotificationGroupBiz) CreateNotificationGroup(ctx context.Context, req *bo.CreateNotificationGroupBo) (snowflake.ID, error) {
	taken, err := b.repo.NotificationGroupNameTaken(ctx, req.Name, 0)
	if err != nil {
		b.helper.Errorw("msg", "check notification group name taken failed", "error", err, "name", req.Name)
		return 0, merr.ErrorInternalServer("check name failed").WithCause(err)
	}
	if taken {
		return 0, merr.ErrorParams("notification group name already exists, please use another name")
	}
	uid, err := b.repo.CreateNotificationGroup(ctx, req)
	if err != nil {
		b.helper.Errorw("msg", "create notification group failed", "error", err, "req", req)
		return 0, merr.ErrorInternalServer("create notification group failed").WithCause(err)
	}
	return uid, nil
}

func (b *NotificationGroupBiz) UpdateNotificationGroup(ctx context.Context, req *bo.UpdateNotificationGroupBo) error {
	taken, err := b.repo.NotificationGroupNameTaken(ctx, req.Name, req.UID)
	if err != nil {
		b.helper.Errorw("msg", "check notification group name taken failed", "error", err, "name", req.Name)
		return merr.ErrorInternalServer("check name failed").WithCause(err)
	}
	if taken {
		return merr.ErrorParams("notification group name already exists, please use another name")
	}
	if err := b.repo.UpdateNotificationGroup(ctx, req); err != nil {
		b.helper.Errorw("msg", "update notification group failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update notification group failed").WithCause(err)
	}
	return nil
}

func (b *NotificationGroupBiz) UpdateNotificationGroupStatus(ctx context.Context, req *bo.UpdateNotificationGroupStatusBo) error {
	if err := b.repo.UpdateNotificationGroupStatus(ctx, req); err != nil {
		b.helper.Errorw("msg", "update notification group status failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update notification group status failed").WithCause(err)
	}
	return nil
}

func (b *NotificationGroupBiz) DeleteNotificationGroup(ctx context.Context, uid snowflake.ID) error {
	if err := b.repo.DeleteNotificationGroup(ctx, uid); err != nil {
		b.helper.Errorw("msg", "delete notification group failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("delete notification group failed").WithCause(err)
	}
	return nil
}

func (b *NotificationGroupBiz) GetNotificationGroup(ctx context.Context, uid snowflake.ID) (*bo.NotificationGroupItemBo, error) {
	item, err := b.repo.GetNotificationGroup(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("notification group %d not found", uid.Int64())
		}
		b.helper.Errorw("msg", "get notification group failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get notification group failed").WithCause(err)
	}
	if err = b.fillMemberProfile(ctx, []*bo.NotificationGroupItemBo{item}); err != nil {
		b.helper.Errorw("msg", "fill notification group member profile failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("fill notification group member profile failed").WithCause(err)
	}
	if err = b.fillNotificationResourceProfile(ctx, []*bo.NotificationGroupItemBo{item}); err != nil {
		b.helper.Errorw("msg", "fill notification group resource profile failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("fill notification group resource profile failed").WithCause(err)
	}
	return item, nil
}

func (b *NotificationGroupBiz) ListNotificationGroup(ctx context.Context, req *bo.ListNotificationGroupBo) (*bo.PageResponseBo[*bo.NotificationGroupItemBo], error) {
	result, err := b.repo.ListNotificationGroup(ctx, req)
	if err != nil {
		b.helper.Errorw("msg", "list notification group failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list notification group failed").WithCause(err)
	}
	if err = b.fillMemberProfile(ctx, result.GetItems()); err != nil {
		b.helper.Errorw("msg", "fill notification group member profile failed", "error", err)
		return nil, merr.ErrorInternalServer("fill notification group member profile failed").WithCause(err)
	}
	if err = b.fillNotificationResourceProfile(ctx, result.GetItems()); err != nil {
		b.helper.Errorw("msg", "fill notification group resource profile failed", "error", err)
		return nil, merr.ErrorInternalServer("fill notification group resource profile failed").WithCause(err)
	}
	return result, nil
}

func (b *NotificationGroupBiz) fillMemberProfile(ctx context.Context, groups []*bo.NotificationGroupItemBo) error {
	if len(groups) == 0 {
		return nil
	}
	uidSet := make(map[int64]struct{})
	for _, group := range groups {
		if group == nil {
			continue
		}
		for _, member := range group.Members {
			if member != nil && member.MemberUID > 0 {
				uidSet[member.MemberUID] = struct{}{}
			}
		}
	}
	if len(uidSet) == 0 {
		return nil
	}
	memberUIDs := make([]int64, 0, len(uidSet))
	for memberUID := range uidSet {
		memberUIDs = append(memberUIDs, memberUID)
	}

	// Keep output deterministic for logs/debugging and stable chunking.
	slices.Sort(memberUIDs)
	const chunkSize = 200
	memberMap := make(map[int64]*goddessv1.MemberItem, len(memberUIDs))
	for i := 0; i < len(memberUIDs); i += chunkSize {
		end := i + chunkSize
		if end > len(memberUIDs) {
			end = len(memberUIDs)
		}
		reply, err := b.memberRepo.ListMember(ctx, &goddessv1.ListMemberRequest{
			Page:     1,
			PageSize: chunkSize,
			Uids:     memberUIDs[i:end],
		})
		if err != nil {
			return err
		}
		for _, item := range reply.GetItems() {
			if item != nil {
				memberMap[item.GetUid()] = item
			}
		}
	}

	for _, group := range groups {
		if group == nil {
			continue
		}
		for _, member := range group.Members {
			if member == nil {
				continue
			}
			item, ok := memberMap[member.MemberUID]
			if !ok || item == nil {
				continue
			}
			member.MemberName = item.GetName()
			member.MemberAvatar = item.GetAvatar()
		}
	}
	return nil
}

func (b *NotificationGroupBiz) fillNotificationResourceProfile(ctx context.Context, groups []*bo.NotificationGroupItemBo) error {
	webhookUIDSet := make(map[int64]struct{})
	templateUIDSet := make(map[int64]struct{})
	emailUIDSet := make(map[int64]struct{})

	for _, group := range groups {
		if group == nil {
			continue
		}
		for _, uid := range group.Webhooks {
			if uid > 0 {
				webhookUIDSet[uid] = struct{}{}
			}
		}
		for _, uid := range group.Templates {
			if uid > 0 {
				templateUIDSet[uid] = struct{}{}
			}
		}
		for _, uid := range group.EmailConfigs {
			if uid > 0 {
				emailUIDSet[uid] = struct{}{}
			}
		}
	}

	eg, egCtx := errgroup.WithContext(ctx)
	var (
		webhookMap  map[int64]*bo.NotificationResourceItemBo
		templateMap map[int64]*bo.NotificationResourceItemBo
		emailMap    map[int64]*bo.NotificationResourceItemBo
	)
	eg.Go(func() error {
		var err error
		webhookMap, err = b.batchListWebhookItems(egCtx, webhookUIDSet)
		return err
	})
	eg.Go(func() error {
		var err error
		templateMap, err = b.batchListTemplateItems(egCtx, templateUIDSet)
		return err
	})
	eg.Go(func() error {
		var err error
		emailMap, err = b.batchListEmailConfigItems(egCtx, emailUIDSet)
		return err
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	for _, group := range groups {
		if group == nil {
			continue
		}
		group.WebhookItems = make([]*bo.NotificationResourceItemBo, 0, len(group.Webhooks))
		for _, uid := range group.Webhooks {
			if item, ok := webhookMap[uid]; ok {
				group.WebhookItems = append(group.WebhookItems, item)
			}
		}

		group.TemplateItems = make([]*bo.NotificationResourceItemBo, 0, len(group.Templates))
		for _, uid := range group.Templates {
			if item, ok := templateMap[uid]; ok {
				group.TemplateItems = append(group.TemplateItems, item)
			}
		}

		group.EmailConfigItems = make([]*bo.NotificationResourceItemBo, 0, len(group.EmailConfigs))
		for _, uid := range group.EmailConfigs {
			if item, ok := emailMap[uid]; ok {
				group.EmailConfigItems = append(group.EmailConfigItems, item)
			}
		}
	}
	return nil
}

func (b *NotificationGroupBiz) batchListWebhookItems(ctx context.Context, uidSet map[int64]struct{}) (map[int64]*bo.NotificationResourceItemBo, error) {
	const (
		chunkSize           = 200
		chunkConcurrencyMax = 4
		cacheTTL            = 30 * time.Second
	)
	nowUnixNano := time.Now().UnixNano()

	result := make(map[int64]*bo.NotificationResourceItemBo, len(uidSet))
	missingUIDs := make([]int64, 0, len(uidSet))

	for uid := range uidSet {
		if uid <= 0 {
			continue
		}
		key := "webhook:" + strconv.FormatInt(uid, 10)
		if name, ok := notificationResourceNameTTLCacheStore.get(key, nowUnixNano); ok {
			result[uid] = &bo.NotificationResourceItemBo{UID: uid, Name: name}
			continue
		}
		missingUIDs = append(missingUIDs, uid)
	}

	if len(missingUIDs) == 0 {
		return result, nil
	}

	eg, egCtx := errgroup.WithContext(ctx)
	var (
		mu sync.Mutex
	)
	sem := make(chan struct{}, chunkConcurrencyMax)

	for i := 0; i < len(missingUIDs); i += chunkSize {
		end := i + chunkSize
		if end > len(missingUIDs) {
			end = len(missingUIDs)
		}
		chunk := append([]int64(nil), missingUIDs[i:end]...)
		eg.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			reply, err := b.rabbitWebhookRepo.ListWebhook(egCtx, &rabbitv1.ListWebhookRequest{
				Page:     1,
				PageSize: int32(chunkSize),
				Uids:     chunk,
			})
			if err != nil {
				return err
			}

			mu.Lock()
			defer mu.Unlock()
			for _, item := range reply.GetItems() {
				if item == nil {
					continue
				}
				uid := item.GetUid()
				name := item.GetName()
				result[uid] = &bo.NotificationResourceItemBo{UID: uid, Name: name}
				cacheKey := "webhook:" + strconv.FormatInt(uid, 10)
				notificationResourceNameTTLCacheStore.set(cacheKey, name, cacheTTL, nowUnixNano)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

func (b *NotificationGroupBiz) batchListTemplateItems(ctx context.Context, uidSet map[int64]struct{}) (map[int64]*bo.NotificationResourceItemBo, error) {
	const (
		chunkSize           = 200
		chunkConcurrencyMax = 4
		cacheTTL            = 30 * time.Second
	)
	nowUnixNano := time.Now().UnixNano()

	result := make(map[int64]*bo.NotificationResourceItemBo, len(uidSet))
	missingUIDs := make([]int64, 0, len(uidSet))
	for uid := range uidSet {
		if uid <= 0 {
			continue
		}
		key := "template:" + strconv.FormatInt(uid, 10)
		if name, ok := notificationResourceNameTTLCacheStore.get(key, nowUnixNano); ok {
			result[uid] = &bo.NotificationResourceItemBo{UID: uid, Name: name}
			continue
		}
		missingUIDs = append(missingUIDs, uid)
	}
	if len(missingUIDs) == 0 {
		return result, nil
	}

	eg, egCtx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, chunkConcurrencyMax)
	var mu sync.Mutex

	for i := 0; i < len(missingUIDs); i += chunkSize {
		end := i + chunkSize
		if end > len(missingUIDs) {
			end = len(missingUIDs)
		}
		chunk := append([]int64(nil), missingUIDs[i:end]...)
		eg.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			reply, err := b.rabbitTemplateRepo.ListTemplate(egCtx, &rabbitv1.ListTemplateRequest{
				Page:     1,
				PageSize: int32(chunkSize),
				Uids:     chunk,
			})
			if err != nil {
				return err
			}
			mu.Lock()
			defer mu.Unlock()
			for _, item := range reply.GetItems() {
				if item == nil {
					continue
				}
				uid := item.GetUid()
				name := item.GetName()
				result[uid] = &bo.NotificationResourceItemBo{UID: uid, Name: name}
				cacheKey := "template:" + strconv.FormatInt(uid, 10)
				notificationResourceNameTTLCacheStore.set(cacheKey, name, cacheTTL, nowUnixNano)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

func (b *NotificationGroupBiz) batchListEmailConfigItems(ctx context.Context, uidSet map[int64]struct{}) (map[int64]*bo.NotificationResourceItemBo, error) {
	const (
		chunkSize           = 200
		chunkConcurrencyMax = 4
		cacheTTL            = 30 * time.Second
	)
	nowUnixNano := time.Now().UnixNano()

	result := make(map[int64]*bo.NotificationResourceItemBo, len(uidSet))
	missingUIDs := make([]int64, 0, len(uidSet))
	for uid := range uidSet {
		if uid <= 0 {
			continue
		}
		key := "email:" + strconv.FormatInt(uid, 10)
		if name, ok := notificationResourceNameTTLCacheStore.get(key, nowUnixNano); ok {
			result[uid] = &bo.NotificationResourceItemBo{UID: uid, Name: name}
			continue
		}
		missingUIDs = append(missingUIDs, uid)
	}
	if len(missingUIDs) == 0 {
		return result, nil
	}

	eg, egCtx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, chunkConcurrencyMax)
	var mu sync.Mutex

	for i := 0; i < len(missingUIDs); i += chunkSize {
		end := i + chunkSize
		if end > len(missingUIDs) {
			end = len(missingUIDs)
		}
		chunk := append([]int64(nil), missingUIDs[i:end]...)
		eg.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			reply, err := b.rabbitEmailRepo.ListEmailConfig(egCtx, &rabbitv1.ListEmailConfigRequest{
				Page:     1,
				PageSize: int32(chunkSize),
				Uids:     chunk,
			})
			if err != nil {
				return err
			}
			mu.Lock()
			defer mu.Unlock()
			for _, item := range reply.GetItems() {
				if item == nil {
					continue
				}
				uid := item.GetUid()
				name := item.GetName()
				result[uid] = &bo.NotificationResourceItemBo{UID: uid, Name: name}
				cacheKey := "email:" + strconv.FormatInt(uid, 10)
				notificationResourceNameTTLCacheStore.set(cacheKey, name, cacheTTL, nowUnixNano)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}
