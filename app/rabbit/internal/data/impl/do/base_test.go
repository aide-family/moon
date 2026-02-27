package do_test

import (
	"testing"

	klog "github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"

	"github.com/aide-family/rabbit/internal/data/impl/do"
)

var genConfig = gen.Config{
	OutPath: "../query",
	Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	// If you want to generate pointer type properties for nullable fields, set FieldNullable to true
	// FieldNullable: true,
	// If you want to assign default values to fields in the `Create` API, set FieldCoverable to true, see: https://gorm.io/docs/create.html#Default-Values
	FieldCoverable: true,
	// If you want to generate unsigned integer type fields, set FieldSignable to true
	FieldSignable: true,
	// If you want to generate index tags from the database, set FieldWithIndexTag to true
	FieldWithIndexTag: true,
	// If you want to generate type tags from the database, set FieldWithTypeTag to true
	FieldWithTypeTag: true,
	// If you need unit tests for query code, set WithUnitTest to true
	// WithUnitTest: true,
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
