package bo

type TeamItem struct {
	TeamId uint32
	Uuid   string
}

type LabelNotices interface {
	GetKey() string
	GetValue() string
	GetReceiverRoutes() []string
}
