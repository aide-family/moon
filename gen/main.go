package main

import (
	_ "embed"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"prometheus-manager/gen/strategy"
)

//go:embed dsn.yml
var dsn string

func main() {
	if dsn == "" {
		panic("dsn is empty")
	}
	g := gen.NewGenerator(gen.Config{
		Mode: gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutContext,
		//FieldNullable: true,
		OutPath:      "../dal/query",
		ModelPkgPath: "../dal/model",
		WithUnitTest: true,

		// generate model global configuration
		FieldNullable: true, // generate pointer when field is nullable
		//FieldCoverable:    true, // generate pointer when field has default value
		FieldWithIndexTag: true, // generate with gorm index tag
		FieldWithTypeTag:  true, // generate with gorm column type tag
	})

	// Initialize a *gorm.DB instance
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic(err)
	}

	g.UseDB(db)

	strategy.GenerateStrategy(g)
	g.Execute()
}
