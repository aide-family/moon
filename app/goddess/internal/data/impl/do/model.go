package do

func Models() []any {
	return []any{
		&User{},
		&OAuth2User{},
		&Namespace{},
		&Member{},
	}
}
