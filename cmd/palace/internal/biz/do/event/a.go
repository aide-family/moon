// Package event is a event package for kratos.
package event

func Models() []any {
	return []any{
		&Realtime{},
		&SendMessageLog{},
	}
}
