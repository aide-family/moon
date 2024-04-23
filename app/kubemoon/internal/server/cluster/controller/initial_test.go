package controller

import (
	"reflect"
	"testing"
	"time"

	"github.com/aide-family/moon/app/kubemoon/internal/server/cluster"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestController_Initial(t *testing.T) {
	type fields struct {
		Client      client.Client
		set         cluster.Set
		confGetter  cluster.ConfigGetter
		builderFunc func(name string) (cluster.Client, error)
		middlewares []HandlerFunc
		handler     *handler
	}
	type args struct {
		c *Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *time.Duration
		wantErr bool
	}{
		// TODO: complete test
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Controller{
				Client:      tt.fields.Client,
				set:         tt.fields.set,
				confGetter:  tt.fields.confGetter,
				builderFunc: tt.fields.builderFunc,
				middlewares: tt.fields.middlewares,
				handler:     tt.fields.handler,
			}
			got, err := r.Initial(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initial() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Initial() got = %v, want %v", got, tt.want)
			}
		})
	}
}
