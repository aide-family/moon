package repository

type Health interface {
	Readiness() error
}
