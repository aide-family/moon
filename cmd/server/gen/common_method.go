package gen

import (
	"context"
	"encoding"

	"github.com/aide-family/moon/pkg/util/types"
	"gorm.io/gen"
	"gorm.io/gorm"
)

var (
	_ encoding.BinaryMarshaler   = (*CommonMethod)(nil)
	_ encoding.BinaryUnmarshaler = (*CommonMethod)(nil)
)

// CommonMethod common method
type CommonMethod struct {
	ID uint32
}

// UnmarshalBinary json unmarshal
func (c *CommonMethod) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary json marshal
func (c *CommonMethod) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// String json string
func (c *CommonMethod) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// Create func
func (c *CommonMethod) Create(ctx context.Context, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(c).Error
}

// Update func
func (c *CommonMethod) Update(ctx context.Context, tx *gorm.DB, conds ...gen.Condition) error {
	return tx.WithContext(ctx).Model(c).Where(conds).Updates(c).Error
}

// Delete func
func (c *CommonMethod) Delete(ctx context.Context, tx *gorm.DB, conds ...gen.Condition) error {
	return tx.WithContext(ctx).Where(conds).Delete(c).Error
}
