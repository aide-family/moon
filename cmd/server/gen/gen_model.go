package gen

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// Run gen gorm gen model code
func Run(datasource string, drive, outputPath string, isBiz bool) {
	if drive == "" || outputPath == "" {
		log.Warnw("err", "参数错误", "datasource", datasource, "drive", drive, "outputPath", outputPath)
		return
	}
	c := &gen.Config{
		OutPath: outputPath,
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
		// 如果你希望为可为null的字段生成属性为指针类型, 设置 FieldNullable 为 true
		//FieldNullable: true,
		// 如果你希望在 `Create` API 中为字段分配默认值, 设置 FieldCoverable 为 true, 参考: https://gorm.io/docs/create.html#Default-Values
		FieldCoverable: true,
		//如果你希望生成无符号整数类型字段, 设置 FieldSignable 为 true
		FieldSignable: true,
		// 如果你希望从数据库生成索引标记, 设置 FieldWithIndexTag 为 true
		FieldWithIndexTag: true,
		// 如果你希望从数据库生成类型标记, 设置 FieldWithTypeTag 为 true
		FieldWithTypeTag: true,
		// 如果你需要对查询代码进行单元测试, 设置 WithUnitTest 为 true
		//WithUnitTest: true,
	}

	g := gen.NewGenerator(*c)

	if datasource != "" {
		gormDB, err := gorm.Open(mysql.Open(datasource))
		if !types.IsNil(err) {
			panic(err)
		}
		g.UseDB(gormDB) // reuse your gorm db
	}

	if isBiz {
		g.ApplyBasic(bizmodel.Models()...)
	} else {
		g.ApplyBasic(model.Models()...)
	}

	// Generate the code
	g.Execute()
}
