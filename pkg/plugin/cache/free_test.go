package cache_test

import (
	"context"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/coocood/freecache"
	"reflect"
	"testing"
	"time"
)

var cli = freecache.NewCache(1024 * 1024 * 10)

func TestNewFreeCache(t *testing.T) {
	type args struct {
		cli *freecache.Cache
	}
	tests := []struct {
		name string
		args args
		want cache.ICacher
	}{
		{
			name: "TestNewFreeCache",
			args: args{
				cli: cli,
			},
			want: cache.NewFreeCache(cli),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cache.NewFreeCache(tt.args.cli); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFreeCache() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultCache_Close(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Test_defaultCache_Close",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			if err := d.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_defaultCache_Dec(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		ctx        context.Context
		key        string
		expiration time.Duration
	}
	decCli := freecache.NewCache(1024 * 1024 * 10)
	cc := cache.NewFreeCache(decCli)
	if err := cc.SetInt64(context.Background(), "test", 10, 120*time.Second); err != nil {
		t.Error(err)
		return
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "Test_defaultCache_Dec",
			fields: fields{
				cli:  decCli,
				keys: make(map[string]struct{}),
			},
			args: args{
				ctx:        context.Background(),
				key:        "test",
				expiration: time.Second,
			},
			want:    9,
			wantErr: false,
		},
		{
			name: "Test_defaultCache_Dec",
			fields: fields{
				cli:  decCli,
				keys: make(map[string]struct{}),
			},
			args: args{
				ctx:        context.Background(),
				key:        "test",
				expiration: time.Second,
			},
			want:    8,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cc
			got, err := d.Dec(tt.args.ctx, tt.args.key, tt.args.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Dec() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultCache_DecMin(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		ctx        context.Context
		key        string
		min        int64
		expiration time.Duration
	}
	decCli := freecache.NewCache(1024 * 1024 * 10)
	cc := cache.NewFreeCache(decCli)
	if err := cc.SetInt64(context.Background(), "test", 10, 120*time.Second); err != nil {
		t.Error(err)
		return
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Test_defaultCache_DecMin",
			fields: fields{
				cli:  decCli,
				keys: make(map[string]struct{}),
			},
			args: args{
				ctx:        context.Background(),
				key:        "test",
				min:        9,
				expiration: time.Second,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Test_defaultCache_DecMin",
			fields: fields{
				cli:  decCli,
				keys: make(map[string]struct{}),
			},
			args: args{
				ctx:        context.Background(),
				key:        "test",
				min:        9,
				expiration: time.Second,
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cc
			got, err := d.DecMin(tt.args.ctx, tt.args.key, tt.args.min, tt.args.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecMin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DecMin() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultCache_DelKeys(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		ctx    context.Context
		prefix string
	}
	decCli := freecache.NewCache(1024 * 1024 * 10)
	cc := cache.NewFreeCache(decCli)
	if err := cc.SetInt64(context.Background(), "test", 10, 120*time.Second); err != nil {
		t.Error(err)
		return
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test_defaultCache_DelKeys",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				ctx:    context.Background(),
				prefix: "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			if err := d.DelKeys(tt.args.ctx, tt.args.prefix); (err != nil) != tt.wantErr {
				t.Errorf("DelKeys() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_defaultCache_Delete(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		in0 context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test_defaultCache_Delete",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				in0: context.Background(),
				key: "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			if err := d.Delete(tt.args.in0, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if _, err := d.Get(tt.args.in0, tt.args.key); err == nil {
				t.Error("Get() should return error")
			}
		})
	}
}

func Test_defaultCache_Exist(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Test_defaultCache_Exist",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				ctx: context.Background(),
				key: "test",
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			got, err := d.Exist(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Exist() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultCache_Get(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test_defaultCache_Get",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				ctx: context.Background(),
				key: "test",
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			got, err := d.Get(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultCache_GetBool(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		in0 context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Test_defaultCache_GetBool",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				in0: context.Background(),
				key: "test",
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			got, err := d.GetBool(tt.args.in0, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetBool() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultCache_GetFloat64(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "Test_defaultCache_GetFloat64",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				ctx: context.Background(),
				key: "test",
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			got, err := d.GetFloat64(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFloat64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetFloat64() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultCache_GetInt64(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "Test_defaultCache_GetInt64",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				ctx: context.Background(),
				key: "test",
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			got, err := d.GetInt64(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetInt64() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type TestObject struct {
	Name string `json:"name"`
}

func (t *TestObject) MarshalBinary() (data []byte, err error) {
	return types.Marshal(t)
}

func (t *TestObject) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, t)
}

func Test_defaultCache_GetObject(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		in0 context.Context
		key string
		obj cache.IObjectSchema
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test_defaultCache_GetObject",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				in0: context.Background(),
				key: "test",
				obj: &TestObject{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			if err := d.GetObject(tt.args.in0, tt.args.key, tt.args.obj); (err != nil) != tt.wantErr {
				t.Errorf("GetObject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_defaultCache_Inc(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		ctx        context.Context
		key        string
		expiration time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "Test_defaultCache_Inc",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				ctx:        context.Background(),
				key:        "test",
				expiration: time.Second,
			},
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			got, err := d.Inc(tt.args.ctx, tt.args.key, tt.args.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("Inc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Inc() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultCache_IncMax(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		ctx        context.Context
		key        string
		max        int64
		expiration time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Test_defaultCache_IncMax",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				ctx:        context.Background(),
				key:        "test",
				max:        10,
				expiration: time.Second,
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			got, err := d.IncMax(tt.args.ctx, tt.args.key, tt.args.max, tt.args.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("IncMax() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IncMax() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultCache_Keys(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		in0    context.Context
		prefix string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Test_defaultCache_Keys",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				in0:    context.Background(),
				prefix: "test",
			},
			want:    []string{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			got, err := d.Keys(tt.args.in0, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("Keys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Keys() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultCache_Set(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		in0        context.Context
		key        string
		value      string
		expiration time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test_defaultCache_Set",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				in0:        context.Background(),
				key:        "test",
				value:      "test",
				expiration: time.Second,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			if err := d.Set(tt.args.in0, tt.args.key, tt.args.value, tt.args.expiration); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_defaultCache_SetBool(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		in0        context.Context
		key        string
		value      bool
		expiration time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test_defaultCache_SetBool",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				in0:        context.Background(),
				key:        "test",
				value:      true,
				expiration: time.Second,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			if err := d.SetBool(tt.args.in0, tt.args.key, tt.args.value, tt.args.expiration); (err != nil) != tt.wantErr {
				t.Errorf("SetBool() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_defaultCache_SetFloat64(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		in0        context.Context
		key        string
		value      float64
		expiration time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test_defaultCache_SetFloat64",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				in0:        context.Background(),
				key:        "test",
				value:      1.0,
				expiration: time.Second,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			if err := d.SetFloat64(tt.args.in0, tt.args.key, tt.args.value, tt.args.expiration); (err != nil) != tt.wantErr {
				t.Errorf("SetFloat64() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_defaultCache_SetInt64(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		ctx        context.Context
		key        string
		value      int64
		expiration time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test_defaultCache_SetInt64",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				ctx:        context.Background(),
				key:        "test",
				value:      1,
				expiration: time.Second,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			if err := d.SetInt64(tt.args.ctx, tt.args.key, tt.args.value, tt.args.expiration); (err != nil) != tt.wantErr {
				t.Errorf("SetInt64() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_defaultCache_SetNX(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		ctx        context.Context
		key        string
		value      string
		expiration time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Test_defaultCache_SetNX",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				ctx:        context.Background(),
				key:        "test",
				value:      "test",
				expiration: time.Second,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			got, err := d.SetNX(tt.args.ctx, tt.args.key, tt.args.value, tt.args.expiration)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetNX() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SetNX() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultCache_SetObject(t *testing.T) {
	type fields struct {
		cli  *freecache.Cache
		keys map[string]struct{}
	}
	type args struct {
		in0        context.Context
		key        string
		obj        cache.IObjectSchema
		expiration time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test_defaultCache_SetObject",
			fields: fields{
				cli:  freecache.NewCache(1024 * 1024 * 10),
				keys: make(map[string]struct{}),
			},
			args: args{
				in0:        context.Background(),
				key:        "test",
				obj:        &TestObject{},
				expiration: time.Second,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := cache.NewFreeCache(tt.fields.cli)
			if err := d.SetObject(tt.args.in0, tt.args.key, tt.args.obj, tt.args.expiration); (err != nil) != tt.wantErr {
				t.Errorf("SetObject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
