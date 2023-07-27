package strategy

import (
	"context"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

type (
	Filter interface {
		// select * from @@table where id = @id
		WhereID(ctx context.Context, id uint) ([]*gen.T, error)
	}
)

const moduleNamePrefix = "strategy_"

func joinModuleName(name string) string {
	return moduleNamePrefix + name
}

func GenerateStrategy(g *gen.Generator) {
	strategiesTableName := "strategies"
	groupsTableName := joinModuleName("groups")
	strategies := g.GenerateModel(strategiesTableName)
	groups := g.GenerateModel(groupsTableName)
	g.ApplyInterface(func(Filter) {}, strategies)
	g.ApplyInterface(func(Filter) {}, groups)
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
}
