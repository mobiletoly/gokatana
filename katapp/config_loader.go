package katapp

import (
	"errors"
	"fmt"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
	"os"
	"reflect"
	"strings"
)

const baseDir = "./configs"

type Deployment struct {
	Name            string
	ConfigDir       string
	CommonConfigDir string
}

// LoadConfig loads configuration from yaml files and environment variables.
// It loads common configuration first and then overrides it with deployment-specific configuration.
// If commonDir is provided (not null), it will be used as a base directory for common configuration, but
// deployment-specific configuration will still be loaded from the default base directory.
func LoadConfig[T any](
	envVarPrefix string,
	deployment Deployment,
) *T {
	var commonBaseDir string
	if deployment.CommonConfigDir == "" {
		commonBaseDir = baseDir
	} else {
		commonBaseDir = deployment.CommonConfigDir
	}
	opt := loadOptions{
		EnvVarPrefix: envVarPrefix,
		Deployment:   "common",
		BaseDir:      commonBaseDir,
	}
	var cfg T
	err := loadFromYaml(opt, &cfg)
	if err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			panic(fmt.Sprintf("failed to load a configuration file: %v", err))
		}
	}
	var deploymentBaseDir string
	if deployment.ConfigDir == "" {
		deploymentBaseDir = baseDir
	} else {
		deploymentBaseDir = deployment.ConfigDir
	}
	opt = loadOptions{
		EnvVarPrefix: envVarPrefix,
		Deployment:   deployment.Name,
		BaseDir:      deploymentBaseDir,
	}
	err = loadFromYaml(opt, &cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration file: %v", err))
	}
	return &cfg
}

type loadOptions struct {
	EnvVarPrefix string
	Deployment   string
	BaseDir      string
}

func loadFromYaml[T any](opt loadOptions, cfg *T) error {
	v := viper.New()
	v.SetEnvPrefix(opt.EnvVarPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetConfigName(opt.Deployment)
	v.SetConfigType("yaml")
	v.AddConfigPath(opt.BaseDir)
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading configuration file: %w", err)
	}
	err := v.Unmarshal(cfg, decoderWithEnvVariablesSupport())
	if err != nil {
		return fmt.Errorf("error parsing configuration file: %w", err)
	}
	return nil
}

// decoderWithEnvVariablesSupport allows us to resolve values containing environment variables,
// e.g. in yaml file you should be able to use constructions such as:
//
//	file_path: "${HOME}/notes.txt"
func decoderWithEnvVariablesSupport() viper.DecoderConfigOption {
	return func(c *mapstructure.DecoderConfig) {
		c.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			c.DecodeHook,
			mapstructure.StringToSliceHookFunc(","),
			replaceEnvVarsHookFunc,
		)
	}
}

func replaceEnvVars(value string) string {
	return os.ExpandEnv(value)
}

func replaceEnvVarsHookFunc(
	f reflect.Type,
	t reflect.Type,
	data interface{},
) (interface{}, error) {
	if f.Kind() != reflect.String || t.Kind() != reflect.String {
		return data, nil
	}

	return replaceEnvVars(data.(string)), nil
}
