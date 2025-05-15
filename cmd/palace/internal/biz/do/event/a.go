package event

func Models() []any {
	return []any{
		&Realtime{},
		&SendMessageLog{},
	}
}
