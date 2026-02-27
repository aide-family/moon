// Package repository is the repository package for the rabbit service.
package repository

type Health interface {
	Readiness() error
}
