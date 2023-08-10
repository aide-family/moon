package buildQuery

import (
	"gorm.io/gen"
	"gorm.io/gen/field"
)

type (
	ISizer interface {
		GetSize() int32
	}
	ICurrenter interface {
		GetCurrent() int32
	}
	IPage interface {
		ISizer
		ICurrenter
	}
	ISort interface {
		GetField() string
		GetAsc() bool
	}
	IField interface {
		GetFields() []string
	}
	ISortGetter interface {
		GetSort() []ISort
	}
	IQuery interface {
		GetPage() IPage
		ISortGetter
		IField
		GetKeyword() string
	}
	IListRequest interface {
		GetQuery()
	}
	IDBQeury interface {
		GetFieldByName(fieldName string) (field.OrderExpr, bool)
	}

	TimeRange struct {
		StartField field.Time
		StartAt    int64
		EndField   field.Time
		EndAt      int64
	}
)

const (
	DefaultOffset int32 = 0
	DefaultLimit  int32 = 10
)

func GetSlectExprs(q IDBQeury, params IField) []field.Expr {
	selectFields := params.GetFields()
	if selectFields != nil && len(selectFields) > 0 {
		selectExpr := make([]field.Expr, 0, len(selectFields))
		for _, fieldName := range selectFields {
			fieldExpr, ok := q.GetFieldByName(fieldName)
			if ok {
				continue
			}
			selectExpr = append(selectExpr, fieldExpr)
		}
		return selectExpr
	}

	return nil
}

func GetSorts(q IDBQeury, sorts ...ISort) []field.Expr {
	sortExpr := make([]field.Expr, 0, len(sorts))
	for _, sort := range sorts {
		fieldExpr, ok := q.GetFieldByName(sort.GetField())
		if ok {
			continue
		}
		if !sort.GetAsc() {
			sortExpr = append(sortExpr, fieldExpr.Desc())
		}
	}

	return sortExpr
}

func GetPage(params IPage) (offset, limit int32) {
	offset = (params.GetCurrent() - 1) * params.GetSize()
	limit = params.GetSize()
	if offset < DefaultOffset {
		offset = DefaultOffset
	}

	if limit < DefaultLimit {
		limit = DefaultLimit
	}

	return
}

func GetKeywords(keyword string, fields ...field.String) []gen.Condition {
	keywordExpr := make([]gen.Condition, 0, len(fields))
	for _, f := range fields {
		keywordExpr = append(keywordExpr, f.Like(keyword))
	}

	return keywordExpr
}
