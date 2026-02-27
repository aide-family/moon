package connect

import (
	"gorm.io/gorm"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/merr"
)

// NewDB creates a new database connection.
// If the orm factory is not registered, it will return an error.
// The database connection is not closed, you need to call the returned function to close the connection.
func NewDB(c *config.ORMConfig) (*gorm.DB, func() error, error) {
	factory, ok := globalRegistry.GetORMFactory(c.GetDialector())
	if !ok {
		return nil, nil, merr.ErrorInternalServer("%s orm factory not registered", c.GetDialector())
	}
	db, err := factory(c)
	if err != nil {
		return nil, nil, err
	}
	return db, func() error { return closeDBConnection(db) }, nil
}

func closeDBConnection(db *gorm.DB) error {
	mdb, err := db.DB()
	if err != nil {
		return err
	}
	return mdb.Close()
}

func RegisterORMFactory(dialector config.ORMConfig_Dialector, factory ORMFactory) {
	globalRegistry.RegisterORMFactory(dialector, factory)
}

func GetORMFactory(dialector config.ORMConfig_Dialector) (ORMFactory, bool) {
	return globalRegistry.GetORMFactory(dialector)
}
