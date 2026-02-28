package do_test

import (
	"os"
	"testing"

	"github.com/glebarez/sqlite"
	klog "github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"

	"github.com/aide-family/marksman/internal/data/impl/do"
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
	klog.Debugw("msg", "remove all files")
	os.RemoveAll(genConfig.OutPath)
	klog.Debugw("msg", "remove all files success", "path", genConfig.OutPath)

	g := gen.NewGenerator(genConfig)

	klog.Debugw("msg", "generate code start")
	g.ApplyBasic(do.Models()...)
	g.Execute()
	klog.Debugw("msg", "generate code success")
}

func migrateMysql() {
	dsn := "root:123456@tcp(localhost:3306)/marksman?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(do.Models()...)
}

func migrateSQLite() error {
	dsn := "file:../../../../marksman.db?cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect database")
	}
	return db.AutoMigrate(do.Models()...)
}

func TestGenerate(t *testing.T) {
	generate()
}

// func TestMigrateMysql(t *testing.T) {
// 	migrateMysql()
// }

func TestMigrateSQLite(t *testing.T) {
	if err := migrateSQLite(); err != nil {
		t.Fatalf("migrate sqlite failed: %v", err)
	}
}
