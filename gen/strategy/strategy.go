package strategy

import (
	"context"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

type (
	Filter interface {
		// select * from @@table where id = @id
		SaFindById(ctx context.Context, id int32) (*gen.T, error)
	}
)

const moduleNamePrefix = "prom_"

func joinModuleName(name string) string {
	return moduleNamePrefix + name
}

func GenerateStrategy(g *gen.Generator) {
	alarmPagesTableName := joinModuleName("alarm_pages")
	dictTableName := joinModuleName("dict")

	promGroupsTableName := joinModuleName("groups")
	promStrategiesTableName := joinModuleName("strategies")

	alarmPagesTable := g.GenerateModel(alarmPagesTableName)
	dictTable := g.GenerateModel(dictTableName)
	promGroupsTable := g.GenerateModel(promGroupsTableName)
	promStrategiesTable := g.GenerateModel(promStrategiesTableName)

	alarmPagesTable = g.GenerateModel(alarmPagesTableName,
		gen.FieldRelate(field.HasMany,
			"PromStrategies",
			promStrategiesTable,
			&field.RelateConfig{
				GORMTag: field.GormTag{
					"many2many":      []string{joinModuleName("strategy_alarm_pages")},
					"foreignKey":     []string{"ID"},
					"joinForeignKey": []string{"AlarmPageID"},
					"References":     []string{"ID"},
					"joinReferences": []string{"PromStrategyID"},
				},
				RelateSlicePointer: true,
			},
		),
	)

	promStrategiesTable = g.GenerateModel(promStrategiesTableName,
		gen.FieldRelate(field.HasMany,
			"AlarmPages",
			alarmPagesTable,
			&field.RelateConfig{
				GORMTag: field.GormTag{
					"many2many":      []string{joinModuleName("strategy_alarm_pages")},
					"foreignKey":     []string{"ID"},
					"joinForeignKey": []string{"PromStrategyID"},
					"References":     []string{"ID"},
					"joinReferences": []string{"AlarmPageID"},
				},
				RelateSlicePointer: true,
			},
		),
		gen.FieldRelate(field.HasMany,
			"Categories",
			dictTable,
			&field.RelateConfig{
				GORMTag: field.GormTag{
					"many2many":      []string{joinModuleName("strategy_categories")},
					"foreignKey":     []string{"ID"},
					"joinForeignKey": []string{"PromStrategyID"},
					"References":     []string{"ID"},
					"joinReferences": []string{"DictID"},
				},
				RelateSlicePointer: true,
			},
		),
		gen.FieldRelate(field.BelongsTo,
			"AlertLevel",
			dictTable,
			&field.RelateConfig{
				GORMTag: field.GormTag{
					"foreignKey": []string{"alert_level_id"},
				},
				RelatePointer: true,
			},
		),
		gen.FieldRelate(field.BelongsTo,
			"GroupInfo",
			promGroupsTable,
			&field.RelateConfig{
				GORMTag: field.GormTag{
					"foreignKey": []string{"group_id"},
				},
				RelatePointer: true,
			},
		),
	)

	promGroupsTable = g.GenerateModel(promGroupsTableName,
		gen.FieldRelate(field.HasMany,
			"PromStrategies",
			promStrategiesTable,
			&field.RelateConfig{
				GORMTag: field.GormTag{
					"foreignKey": []string{"GroupID"},
				},
				RelateSlicePointer: true,
			},
		),
		gen.FieldRelate(field.HasMany,
			"Categories",
			dictTable,
			&field.RelateConfig{
				GORMTag: field.GormTag{
					"many2many":      []string{joinModuleName("group_categories")},
					"foreignKey":     []string{"ID"},
					"joinForeignKey": []string{"PromGroupID"},
					"References":     []string{"ID"},
					"joinReferences": []string{"DictID"},
				},
				RelateSlicePointer: true,
			},
		),
	)

	g.ApplyInterface(func(Filter) {}, alarmPagesTable)
	g.ApplyInterface(func(Filter) {}, promStrategiesTable)
	g.ApplyInterface(func(Filter) {}, promGroupsTable)
	g.ApplyInterface(func(Filter) {}, dictTable)
}
