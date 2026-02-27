// Package orm provides a GORM builder.
package orm

import "gorm.io/gorm"

// Builder is a GORM builder.
type Builder struct {
	Dialector gorm.Dialector
	Config    *gorm.Config
	IsDebug   bool
}

// Build builds a new GORM database connection.
// If the GORM database connection is not valid, it will return an error.
func (c *Builder) Build() (*gorm.DB, error) {
	db, err := gorm.Open(c.Dialector, c.Config)
	if err != nil {
		return nil, err
	}

	if c.IsDebug {
		return db.Debug(), nil
	}

	return db, nil
}
