package gen

import (
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

func Run(datasource string, drive, outputPath string) {
	if drive == "" || outputPath == "" || datasource == "" {
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

	gormDB, err := gorm.Open(mysql.Open(datasource))
	if err != nil {
		panic(err)
	}
	g.UseDB(gormDB) // reuse your gorm db
	g.WithOpts(gen.WithMethod(new(CommonMethod).String))
	g.WithOpts(gen.WithMethod(new(CommonMethod).UnmarshalBinary))
	g.WithOpts(gen.WithMethod(new(CommonMethod).MarshalBinary))
	g.WithOpts(gen.WithMethod(new(CommonMethod).Create))
	g.WithOpts(gen.WithMethod(new(CommonMethod).Update))
	g.WithOpts(gen.WithMethod(new(CommonMethod).Delete))

	g.WithOpts(gen.FieldType("id", "uint32"),
		gen.FieldType("gender", "int"),
		gen.FieldType("status", "int"),
		gen.FieldType("role", "int"),
		//gen.FieldType("updated_at", "*types.Time"),
		//gen.FieldType("created_at", "*types.Time"),
	)

	usersTable := g.GenerateModel("sys_users")
	teamTable := g.GenerateModel("sys_teams")
	apiTable := g.GenerateModel("sys_apis")
	sysTeamRoleTable := g.GenerateModel("sys_team_roles")
	// sys_team_member_roles
	sysTeamMemberRoleTable := g.GenerateModel("sys_team_member_roles")
	// casbin_rule
	casbinRuleTable := g.GenerateModel("casbin_rule")

	//tables := g.GenerateAllTable()
	var tables []any
	tables = append(tables, g.GenerateModel("sys_teams",
		gen.FieldRelate(field.HasOne, "Leader", usersTable, &field.RelateConfig{
			GORMTag: field.GormTag{
				"foreignKey": []string{"LeaderID"},
			},
			RelatePointer: true,
		}),
		gen.FieldRelate(field.HasOne, "Creator", usersTable, &field.RelateConfig{
			GORMTag: field.GormTag{
				"foreignKey": []string{"CreatorID"},
			},
			RelatePointer: true,
		}),
	), g.GenerateModel("sys_team_members",
		gen.FieldRelate(field.HasOne, "Member", usersTable, &field.RelateConfig{
			GORMTag: field.GormTag{
				"foreignKey": []string{"UserID"},
			},
			RelatePointer: true,
		}),
		gen.FieldRelate(field.HasOne, "Team", teamTable, &field.RelateConfig{
			GORMTag: field.GormTag{
				"foreignKey": []string{"TeamID"},
			},
			RelatePointer: true,
		}),
		gen.FieldRelate(field.Many2Many, "TeamRoles", sysTeamRoleTable, &field.RelateConfig{
			GORMTag: field.GormTag{
				"many2many": []string{"sys_team_member_roles"},
			},
			RelateSlicePointer: true,
		}),
	), g.GenerateModel("sys_team_roles",
		gen.FieldRelate(field.Many2Many, "Apis", apiTable, &field.RelateConfig{
			GORMTag: field.GormTag{
				"many2many": []string{"sys_team_role_apis"},
			},
			RelateSlicePointer: true,
		}),
	), g.GenerateModel("sys_apis",
		gen.FieldRelate(field.Many2Many, "SysTeamRoles", sysTeamRoleTable, &field.RelateConfig{
			GORMTag: field.GormTag{
				"many2many": []string{"sys_team_role_apis"},
			},
			RelateSlicePointer: true,
		}),
	), sysTeamMemberRoleTable, usersTable, casbinRuleTable)

	g.ApplyBasic(tables...)

	// Generate the code
	g.Execute()
}
