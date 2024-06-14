package manager

import (
	"context"
	"github.com/aide-family/moon/pkg/rabbit"
	"github.com/aide-family/moon/pkg/rabbit/template"
)

var _ rabbit.TemplaterGetter = &TemplateManger{}

type TemplateManger struct {
}

func (t *TemplateManger) Get(context context.Context, id int64) (rabbit.Templater, error) {
	//TODO implement me
	var templateContent string
	return &template.GenericTemplateParser{Template: templateContent}, nil
}

func (t *TemplateManger) GetSecret(context context.Context, id int64) ([]byte, error) {
	//TODO implement me
	var secret []byte
	return secret, nil
}

func (t *TemplateManger) GetSuppressorID(context context.Context, id int64) (int64, error) {
	//TODO implement me
	var suppressorID int64
	return suppressorID, nil
}

func (t *TemplateManger) GetSenderName(context context.Context, id int64) (string, error) {
	//TODO implement me
	var senderName string
	return senderName, nil
}
