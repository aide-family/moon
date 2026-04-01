//go:build generate

package do_test

import (
	"testing"

	klog "github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"

	"github.com/aide-family/jade_tree/internal/data/impl/do"
)

var genConfig = gen.Config{
	OutPath:        "../query",
	Mode:           gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	FieldCoverable: true,
	FieldSignable:  true,
	FieldWithIndexTag: true,
	FieldWithTypeTag:  true,
}

func generate() {
	g := gen.NewGenerator(genConfig)

	klog.Debugw("msg", "generate code start")
	g.ApplyBasic(do.Models()...)
	g.Execute()
	klog.Debugw("msg", "generate code success")
}

func TestGenerate(t *testing.T) {
	generate()
}
