package strategy

import (
	"context"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

type (
	Filter interface {
		// select * from @@table where id = @id
		WhereID(ctx context.Context, id uint) (*gen.T, error)
	}
)

const moduleNamePrefix = "prom_"

func joinModuleName(name string) string {
	return moduleNamePrefix + name
}

func GenerateStrategy(g *gen.Generator) {
	promNodesTableName := joinModuleName("nodes")
	nodeDirsTableName := joinModuleName("node_dirs")
	filesTableName := joinModuleName("node_dir_files")
	groupsTableName := joinModuleName("node_dir_file_groups")
	strategiesTableName := joinModuleName("node_dir_file_group_strategies")
	combosTableName := joinModuleName("combos")
	rulesTableName := joinModuleName("rules")
	comboStrategiesTableName := joinModuleName("combo_strategies")

	strategies := g.GenerateModel(strategiesTableName)
	groups := g.GenerateModel(groupsTableName)
	files := g.GenerateModel(filesTableName)
	nodeDirs := g.GenerateModel(nodeDirsTableName)
	promNodes := g.GenerateModel(promNodesTableName)
	combos := g.GenerateModel(combosTableName)
	g.GenerateModel(comboStrategiesTableName)
	rules := g.GenerateModel(rulesTableName)

	g.ApplyInterface(func(Filter) {}, strategies)
	g.ApplyInterface(func(Filter) {}, groups)
	g.ApplyInterface(func(Filter) {}, files)
	g.ApplyInterface(func(Filter) {}, nodeDirs)
	g.ApplyInterface(func(Filter) {}, promNodes)
	g.ApplyInterface(func(Filter) {}, combos)
	g.ApplyInterface(func(Filter) {}, rules)

	g.GenerateModel(groupsTableName,
		gen.FieldRelate(field.HasMany,
			"Strategies",
			strategies,
			&field.RelateConfig{
				GORMTag: field.GormTag{
					"foreignKey": []string{"GroupID"},
				},
				RelateSlicePointer: true,
			},
		),
	)

	g.GenerateModel(filesTableName,
		gen.FieldRelate(field.HasMany,
			"Groups",
			groups,
			&field.RelateConfig{
				GORMTag: field.GormTag{
					"foreignKey": []string{"FileID"},
				},
				RelateSlicePointer: true,
			},
		),
	)

	g.GenerateModel(nodeDirsTableName,
		gen.FieldRelate(field.HasMany,
			"Files",
			files,
			&field.RelateConfig{
				GORMTag: field.GormTag{
					"foreignKey": []string{"DirID"},
				},
				RelateSlicePointer: true,
			},
		),
	)

	g.GenerateModel(promNodesTableName,
		gen.FieldRelate(field.HasMany,
			"NodeDirs",
			nodeDirs,
			&field.RelateConfig{
				GORMTag: field.GormTag{
					"foreignKey": []string{"NodeID"},
				},
				RelateSlicePointer: true,
			},
		),
	)

	g.GenerateModel(combosTableName,
		gen.FieldRelate(field.Many2Many,
			"Rules",
			rules,
			&field.RelateConfig{
				GORMTag: field.GormTag{
					"foreignKey": []string{"ComboID"},
				},
				RelateSlicePointer: true,
			},
		),
	)

	g.GenerateModel(rulesTableName,
		gen.FieldRelate(field.Many2Many,
			"Combos",
			combos,
			&field.RelateConfig{
				GORMTag: field.GormTag{
					"foreignKey": []string{"RuleID"},
				},
				RelateSlicePointer: true,
			},
		),
	)
}
