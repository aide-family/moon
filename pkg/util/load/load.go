package load

import (
	"os"
	"path/filepath"
	"regexp"

	"buf.build/go/protoyaml"
	"github.com/joho/godotenv"
	"google.golang.org/protobuf/proto"
)

func Load(cfgPath string, bootstrap proto.Message) error {
	_ = godotenv.Load()

	return filepath.Walk(cfgPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml") {
			yamlBytes, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			err = protoyaml.UnmarshalOptions{
				Path: cfgPath,
			}.Unmarshal(resolveEnv(yamlBytes), bootstrap)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func resolveEnv(content []byte) []byte {
	regex := regexp.MustCompile(`\$\{(\w+)(?::([^}]+))?}`)

	return regex.ReplaceAllFunc(content, func(match []byte) []byte {
		matches := regex.FindSubmatch(match)
		envKey := string(matches[1])
		var defaultValue string

		if len(matches) > 2 {
			defaultValue = string(matches[2])
		}

		if value, exists := os.LookupEnv(envKey); exists {
			return []byte(value)
		}
		return []byte(defaultValue)
	})
}
