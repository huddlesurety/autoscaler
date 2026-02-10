package config

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	MetricIntervalSecs int `mapstructure:"METRIC_INTERVAL_SECS"`
	MetricInterval     time.Duration
	ScaleIntervalSecs  int `mapstructure:"SCALE_INTERVAL_SECS"`
	ScaleInterval      time.Duration

	Railway  RailwayConfig  `mapstructure:"RAILWAY"`
	Temporal TemporalConfig `mapstructure:"TEMPORAL"`
}

func New() (*Config, error) {
	v := viper.NewWithOptions(viper.ExperimentalBindStruct())
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal environment config: %w", err)
	}

	cfg.MetricInterval = time.Second * time.Duration(cfg.MetricIntervalSecs)
	cfg.ScaleInterval = time.Second * time.Duration(cfg.ScaleIntervalSecs)

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	missingVars := validateStruct(reflect.ValueOf(c).Elem(), "")

	if len(missingVars) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missingVars, ", "))
	}

	return nil
}

// validateStruct recursively validates struct fields and returns missing env variable names.
// prefix is used to build nested env variable names (e.g., "DB" -> "DB_HOST").
func validateStruct(val reflect.Value, prefix string) []string {
	typ := val.Type()
	var missingVars []string

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		envName, hasTag := field.Tag.Lookup("mapstructure")
		if !hasTag {
			continue
		}

		fullEnvName := envName
		if prefix != "" {
			fullEnvName = prefix + "_" + envName
		}

		// Handle nested structs recursively
		if fieldValue.Kind() == reflect.Struct {
			// skip validation of external types
			if field.Type.PkgPath() != "" && !strings.HasPrefix(field.Type.PkgPath(), "github.com/huddlesurety") {
				continue
			}
			// Recurse into nested config structs
			missingVars = append(missingVars, validateStruct(fieldValue, fullEnvName)...)
			continue
		}

		if isEmpty(fieldValue) {
			missingVars = append(missingVars, fullEnvName)
		}
	}

	return missingVars
}

func isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		// Bools are never considered "empty" - false is a valid value
		return false
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	default:
		return false
	}
}
