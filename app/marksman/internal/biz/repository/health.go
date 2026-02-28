// Package repository is the repository package for the marksman service.
package repository

type Health interface {
	Readiness() error
}
