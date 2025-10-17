package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pavletto/altituder/internal/elevation"
	"github.com/spf13/cobra"
)

// Config holds application configuration
type Config struct {
	CacheDir      string
	URLTemplate   string
	Subdomains    []string
	DefaultZoom   int
	MaxNativeZoom int
	HeightFactor  float64
	NoDataValues  string
}

// LoadConfig loads configuration from environment variables and command flags
// Flags take precedence over environment variables
func LoadConfig(cmd *cobra.Command) Config {
	cfg := Config{}

	// Load from env or flags (flags take precedence)
	cfg.CacheDir = getConfigString(cmd, "cache-dir", "DDM_CACHE_DIR", "./cache")
	cfg.URLTemplate = getConfigString(cmd, "url-template", "DDM_URL_TEMPLATE", "https://{s}.geodata.microavia.com/srtm/{z}/{y}/{x}.ddm")
	cfg.Subdomains = strings.Split(getConfigString(cmd, "subdomains", "DDM_SUBDOMAINS", "a,b,c"), ",")
	cfg.DefaultZoom = getConfigInt(cmd, "zoom", "DDM_DEFAULT_Z", 14)
	cfg.MaxNativeZoom = getConfigInt(cmd, "max-native-zoom", "DDM_MAX_NATIVE_Z", 14)
	cfg.HeightFactor = getConfigFloat(cmd, "height-factor", "DDM_HEIGHT_FACTOR", 1.0)
	cfg.NoDataValues = getConfigString(cmd, "nodata-values", "DDM_NODATA_CSV", "")

	return cfg
}

// CreateStore creates a new DDM store from the configuration
func (c *Config) CreateStore() (*elevation.Store, error) {
	cfg := elevation.StoreConfig{
		CacheDir:          c.CacheDir,
		URLTemplate:       c.URLTemplate,
		Subdomains:        c.Subdomains,
		PermitDownload:    c.URLTemplate != "",
		HTTPClientTimeout: 15 * time.Second,
		DefaultZoom:       c.DefaultZoom,
		MaxNativeZoom:     c.MaxNativeZoom,
		HeightFactor:      float32(c.HeightFactor),
		NoDataValues:      elevation.ParseNoData(c.NoDataValues),
	}

	return elevation.NewStore(cfg)
}

// getConfigString gets a string value from flag, then env, then default
func getConfigString(cmd *cobra.Command, flagName, envName, defaultValue string) string {
	// Check if flag was explicitly set
	if cmd.Flags().Changed(flagName) {
		val, _ := cmd.Flags().GetString(flagName)
		return val
	}

	// Check environment variable
	if v := os.Getenv(envName); v != "" {
		return v
	}

	// Use default
	return defaultValue
}

// getConfigInt gets an int value from flag, then env, then default
func getConfigInt(cmd *cobra.Command, flagName, envName string, defaultValue int) int {
	// Check if flag was explicitly set
	if cmd.Flags().Changed(flagName) {
		val, _ := cmd.Flags().GetInt(flagName)
		return val
	}

	// Check environment variable
	if v := os.Getenv(envName); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}

	// Use default
	return defaultValue
}

// getConfigFloat gets a float64 value from flag, then env, then default
func getConfigFloat(cmd *cobra.Command, flagName, envName string, defaultValue float64) float64 {
	// Check if flag was explicitly set
	if cmd.Flags().Changed(flagName) {
		val, _ := cmd.Flags().GetFloat64(flagName)
		return val
	}

	// Check environment variable
	if v := os.Getenv(envName); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}

	// Use default
	return defaultValue
}
