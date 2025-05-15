package do

// UserConfigTable represents user table configuration interface
type UserConfigTable interface {
	Creator
	GetTableKey() string
	GetPageSize() int
	GetColumns() []string
}
