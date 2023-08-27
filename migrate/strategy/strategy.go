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
	alarmHistoryTableName := joinModuleName("alarm_histories")
	alarmPageHistoryTableName := joinModuleName("alarm_page_histories")

	alarmPagesTable := g.GenerateModel(alarmPagesTableName)
	dictTable := g.GenerateModel(dictTableName)
	promGroupsTable := g.GenerateModel(promGroupsTableName)
	promStrategiesTable := g.GenerateModel(promStrategiesTableName)

	alarmPageHistoryTable := g.GenerateModel(alarmPageHistoryTableName)
	alarmHistoryTable := g.GenerateModel(alarmHistoryTableName,
		gen.FieldRelate(field.Many2Many,
			"Pages",
			alarmPagesTable,
			&field.RelateConfig{
				RelateSlicePointer: true,
				GORMTag: field.GormTag{
					"many2many":      []string{joinModuleName("prom_alarm_page_histories")},
					"foreignKey":     []string{"ID"},
					"joinForeignKey": []string{"AlarmPageID"},
					"References":     []string{"ID"},
					"joinReferences": []string{"PageID"},
				},
			},
		))

	alarmPagesTable = g.GenerateModel(alarmPagesTableName,
		gen.FieldRelate(field.Many2Many,
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
		gen.FieldRelate(field.Many2Many,
			"Histories",
			alarmHistoryTable,
			&field.RelateConfig{
				RelateSlicePointer: true,
				GORMTag: field.GormTag{
					"many2many":      []string{joinModuleName("alarm_page_histories")},
					"References":     []string{"ID"},
					"joinReferences": []string{"HistoryID"},
					"foreignKey":     []string{"ID"},
					"joinForeignKey": []string{"AlarmPageID"},
				},
			},
		),
	)

	promStrategiesTable = g.GenerateModel(promStrategiesTableName,
		gen.FieldRelate(field.Many2Many,
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
		gen.FieldRelate(field.Many2Many,
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
					"foreignKey": []string{"AlertLevelID"},
				},
				RelatePointer: true,
			},
		),
		gen.FieldRelate(field.BelongsTo,
			"GroupInfo",
			promGroupsTable,
			&field.RelateConfig{
				GORMTag: field.GormTag{
					"foreignKey": []string{"GroupID"},
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
		gen.FieldRelate(field.Many2Many,
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
	g.ApplyInterface(func(Filter) {}, alarmHistoryTable)
	g.ApplyInterface(func(Filter) {}, alarmPageHistoryTable)
}
