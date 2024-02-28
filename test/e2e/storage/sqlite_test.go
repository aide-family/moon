package e2e

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"testing"
)

type Config struct {
	Data struct {
		Database struct {
			Driver string `yaml:"driver"`
			Source string `yaml:"source"`
			Debug  bool   `yaml:"debug"`
		} `yaml:"database"`
	} `yaml:"data"`
}

type Product struct {
	gorm.Model
	Code  string
	Price int
}

func TestSqlite(t *testing.T) {
	exePath, err := os.Getwd()
	assert.NoError(t, err)
	configDir := filepath.Dir(filepath.Dir(filepath.Dir(exePath)))
	configPath := filepath.Join(configDir, "test/fixtures/testdata/storage/config.yaml")
	t.Logf("ConfigPath: %s", configPath)
	yamlFile, err := os.ReadFile(configPath)
	assert.NoError(t, err)

	var conf Config
	err = yaml.Unmarshal(yamlFile, &conf)
	assert.NoError(t, err)

	driver := conf.Data.Database.Driver
	source := conf.Data.Database.Source
	debug := conf.Data.Database.Debug
	t.Logf("Driver: %s", driver)
	t.Logf("Source: %s", source)
	t.Logf("Debug: %v", debug)

	if "sqlite" != driver {
		return
	}

	var opts []gorm.Option
	var dialector gorm.Dialector
	dialector = sqlite.Open(source)
	db, err := gorm.Open(dialector, opts...)
	assert.NoError(t, err)
	if debug {
		db = db.Debug()
	}

	err = db.AutoMigrate(&Product{})
	assert.NoError(t, err)
	db.Create(&Product{Code: "D42", Price: 100})

	var product Product
	db.First(&product, 1)
	assert.Equal(t, "D42", product.Code)
	assert.Equal(t, 100, product.Price)

	db.First(&product, "code = ?", "D42")
	assert.Equal(t, "D42", product.Code)
	assert.Equal(t, 100, product.Price)

	db.Model(&product).Update("Price", 200)
	db.First(&product, 1)
	assert.Equal(t, "D42", product.Code)
	assert.Equal(t, 200, product.Price)

	db.Model(&product).Updates(Product{Price: 200, Code: "F42"})
	db.First(&product, 1)
	assert.Equal(t, "F42", product.Code)
	assert.Equal(t, 200, product.Price)

	db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})
	db.First(&product, 1)
	assert.Equal(t, "F42", product.Code)
	assert.Equal(t, 200, product.Price)

	db.Delete(&product, 1)
	defer func(name string) {
		err := os.Remove(name)
		assert.NoError(t, err)
	}(source)
}
