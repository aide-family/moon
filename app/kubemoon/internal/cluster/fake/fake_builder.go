package fake

import clu "github.com/aide-family/moon/app/kubemoon/internal/cluster"

var _ clu.Builder = &fakeBuilder{}

type fakeBuilder struct{}

func (f fakeBuilder) Complete() (clu.Client, error) {
	return fakeClient{name: name}, nil
}
