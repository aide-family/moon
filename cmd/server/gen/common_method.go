package gen

import (
	"context"
	"encoding"
	"encoding/json"

	"gorm.io/gen"
	"gorm.io/gorm"
)

var _ encoding.BinaryMarshaler = (*CommonMethod)(nil)
var _ encoding.BinaryUnmarshaler = (*CommonMethod)(nil)

// CommonMethod common method
type CommonMethod struct {
	ID uint32
}

// UnmarshalBinary json unmarshal
func (c *CommonMethod) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary json marshal
func (c *CommonMethod) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// String json string
func (c *CommonMethod) String() string {
	bs, _ := json.Marshal(c)
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
