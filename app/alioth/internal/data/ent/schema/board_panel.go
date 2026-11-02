package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// BoardPanel 是大盘里的一个 Grafana 面板。宽度按 12 栏栅格计算，高度按像素，
// sort 决定同一大盘内的展示顺序。
type BoardPanel struct {
	ent.Schema
}

func (BoardPanel) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (BoardPanel) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("board_uid").NotEmpty(),
		field.String("title").NotEmpty(),
		field.String("description").Default(""),
		field.String("grafana_url").NotEmpty(),
		// width 是面板占 12 栏栅格中的列数。
		field.Int("width").Default(6).Range(1, 12),
		// height 是面板高度，单位像素。
		field.Int("height").Default(320).Range(120, 1600),
		field.Int("sort").Default(0),
	}
}

func (BoardPanel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "board_uid"),
		index.Fields("board_uid", "sort").Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}
