// Package repository is the repository package for the Goddess service.
package repository

type Health interface {
	Readiness() error
}
