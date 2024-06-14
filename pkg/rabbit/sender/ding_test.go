package sender

import (
	"encoding/json"
	"testing"
)

func initTestDingSecret() []byte {
	secret := &DingSecret{
		Token:  "testtoken",
		Secret: "testsecret",
	}
	bytes, err := json.Marshal(secret)
	if err != nil {
		return nil
	}
	return bytes
}

func TestDingSecretProvider_Provider(t *testing.T) {
	secret := initTestDingSecret()
	type args struct {
		in  []byte
		out any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Normal case",
			args: args{
				in:  secret,
				out: &DingSecret{},
			},
			wantErr: false,
		},
		{
			name: "Type Error case",
			args: args{
				in:  secret,
				out: &Ding{},
			},
			wantErr: true,
		},
		{
			name: "Secret Error case",
			args: args{
				in:  []byte("test"),
				out: &DingSecret{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DingSecretProvider{}
			if err := d.Provider(tt.args.in, tt.args.out); (err != nil) != tt.wantErr {
				t.Errorf("Provider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
