package lock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"

	"github.com/aide-family/Alioth/pkg/merr"
)

var (
	// ErrNotAcquired means another holder already owns the key.
	ErrNotAcquired = kerrors.New(409, "LOCK_NOT_ACQUIRED", "lock not acquired")
	// ErrTokenMismatch means the key value is not this holder's token.
	ErrTokenMismatch = kerrors.New(409, "LOCK_TOKEN_MISMATCH", "lock token mismatch")
)

const renewScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0
`

const releaseScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
`

// Client acquires distributed locks. Redis is used only for these locks.
type Client struct {
	rdb        *redis.Client
	instanceID string
}

// Hold is one successful acquisition.
type Hold struct {
	client *Client
	key    string
	token  string
}

// New binds a lock client to a Redis connection and an instance id.
func New(rdb *redis.Client, instanceID string) *Client {
	if instanceID == "" {
		instanceID = randomToken()
	}
	return &Client{rdb: rdb, instanceID: instanceID}
}

// Acquire runs SET key token NX PX ttl. A failed Redis call is returned as an error.
func (c *Client) Acquire(ctx context.Context, key string, ttl time.Duration) (*Hold, error) {
	if ttl <= 0 {
		return nil, merr.Invalid("lock ttl must be positive")
	}
	token := c.instanceID + ":" + randomToken()
	res, err := c.rdb.Do(ctx, "SET", key, token, "NX", "PX", ttl.Milliseconds()).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrNotAcquired
	}
	if err != nil {
		log.NewHelper(log.DefaultLogger).Errorf("lock: %v", err)
		return nil, merr.Server("lock failed")
	}
	if res != "OK" {
		return nil, ErrNotAcquired
	}
	return &Hold{client: c, key: key, token: token}, nil
}

// Key returns the locked key.
func (h *Hold) Key() string { return h.key }

// Token returns the holder token.
func (h *Hold) Token() string { return h.token }

// Renew extends the TTL when the stored value still matches this token.
func (h *Hold) Renew(ctx context.Context, ttl time.Duration) error {
	if ttl <= 0 {
		return merr.Invalid("lock ttl must be positive")
	}
	return h.eval(ctx, renewScript, ttl)
}

// Release deletes the key when the stored value still matches this token.
func (h *Hold) Release(ctx context.Context) error {
	return h.eval(ctx, releaseScript, 0)
}

func (h *Hold) eval(ctx context.Context, script string, ttl time.Duration) error {
	args := []any{h.token}
	if script == renewScript {
		args = append(args, ttl.Milliseconds())
	}
	n, err := h.client.rdb.Eval(ctx, script, []string{h.key}, args...).Int64()
	if err != nil {
		log.NewHelper(log.DefaultLogger).Errorf("lock: %v", err)
		return merr.Server("lock failed")
	}
	if n == 0 {
		return ErrTokenMismatch
	}
	return nil
}

func randomToken() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
