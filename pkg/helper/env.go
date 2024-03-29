package helper

const (
	Dev  = "dev"
	Test = "test"
)

func IsDev(env string) bool {
	return env == Dev || env == Test
}
