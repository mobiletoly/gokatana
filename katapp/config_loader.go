package katapp

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
)

const baseDir = "./configs"

type Deployment struct {
	Name            string
	ConfigDir       string
	CommonConfigDir string
}

func LoadConfig[T any](envVarPrefix string, deployment Deployment) *T {
	return LoadConfigWithMerger[T](envVarPrefix, deployment, nil)
}

// LoadConfigWithMerger loads configuration from yaml files and environment variables.
// It loads common configuration first and then overrides it with deployment-specific configuration.
// If commonDir is provided (not null), it will be used as a base directory for common configuration, but
// deployment-specific configuration will still be loaded from the default base directory.
func LoadConfigWithMerger[T any](
	envVarPrefix string,
	deployment Deployment,
	merger func() map[string]any,
) *T {
	{
		var commonBaseDir string
		if deployment.CommonConfigDir == "" {
			commonBaseDir = baseDir
		} else {
			commonBaseDir = deployment.CommonConfigDir
		}

		var deploymentBaseDir string
		if deployment.ConfigDir == "" {
			deploymentBaseDir = baseDir
		} else {
			deploymentBaseDir = deployment.ConfigDir
		}

		v := viper.New()
		v.SetEnvPrefix(envVarPrefix)
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()
		v.SetConfigType("yaml")

		// 1) Load "common" (optional)
		v.SetConfigName("common")
		v.AddConfigPath(commonBaseDir)
		if err := v.ReadInConfig(); err != nil {
			var nf viper.ConfigFileNotFoundError
			if !errors.As(err, &nf) {
				log.Fatalf("failed to load common configuration file: %v", err)
			}
		}

		// 2) Merge deployment-specific overrides
		v.SetConfigName(deployment.Name)
		v.AddConfigPath(deploymentBaseDir)
		if err := v.MergeInConfig(); err != nil {
			log.Fatalf("failed to load deployment configuration file: %v", err)
		}

		// 3) Apply merger-provided overrides (highest precedence via v.Set)
		if merger != nil {
			if merged := (merger)(); merged != nil {
				applyMergeAsOverrides(v, envVarPrefix, merged)
			}
		}

		// 4) Decode into the target struct with env-var expansion support
		var cfg T
		if err := v.Unmarshal(&cfg, decoderWithEnvVariablesSupport()); err != nil {
			log.Fatalf("error parsing configuration: %v", err)
		}

		return &cfg
	}
}

func applyMergeAsOverrides(v *viper.Viper, envPrefix string, m map[string]any) {
	flattenAndSetWithNorm(v, "", envPrefix, m)
}

func flattenAndSetWithNorm(v *viper.Viper, prefix, envPrefix string, val any) {
	switch t := val.(type) {
	case map[string]any:
		for k, v2 := range t {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			flattenAndSetWithNorm(v, key, envPrefix, v2)
		}
	default:
		// If this is a leaf and the key looks env-like, normalize it.
		key := prefix
		if !strings.Contains(key, ".") || strings.Contains(key, "_") {
			key = normalizeKey(key, envPrefix)
		}
		v.Set(key, t) // Set = highest precedence: Set > env > config
	}
}

func normalizeKey(k, envPrefix string) string {
	if k == "" {
		return k
	}
	up := strings.ToUpper(k)
	pfx := strings.ToUpper(envPrefix)
	// Strip "PREFIX_" if present (case-insensitive)
	if pfx != "" && strings.HasPrefix(up, pfx+"_") {
		k = k[len(envPrefix)+1:]
	}
	// If caller already used dotted keys, leave as-is
	if strings.Contains(k, ".") {
		return strings.ToLower(k)
	}
	// Convert underscores to dots (ENV form -> config path)
	k = strings.ReplaceAll(k, "_", ".")
	return strings.ToLower(k)
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
