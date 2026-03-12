package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"picoclaw/agent/pkg/config"
)

const Logo = "🦞"

var (
	version   = "dev"
	gitCommit string
	buildTime string
	goVersion string
)

func GetConfigPath() string {
	// Priority: PICOCLAW_CONFIG > PICOCLAW_HOME/config.json > ~/.picoclaw/config.json
	if configPath := os.Getenv("PICOCLAW_CONFIG"); configPath != "" {
		return configPath
	}
	
	// Check PICOCLAW_HOME environment variable
	if picoclawHome := os.Getenv("PICOCLAW_HOME"); picoclawHome != "" {
		return filepath.Join(picoclawHome, "config.json")
	}
	
	// Default: ~/.picoclaw/config.json
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".picoclaw", "config.json")
}

// GetPicoclawHome returns the PICOCLAW_HOME directory
// Priority: PICOCLAW_HOME env > ~/.picoclaw
func GetPicoclawHome() string {
	if picoclawHome := os.Getenv("PICOCLAW_HOME"); picoclawHome != "" {
		return picoclawHome
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".picoclaw")
}

func LoadConfig() (*config.Config, error) {
	return config.LoadConfig(GetConfigPath())
}

// FormatVersion returns the version string with optional git commit
func FormatVersion() string {
	v := version
	if gitCommit != "" {
		v += fmt.Sprintf(" (git: %s)", gitCommit)
	}
	return v
}

// FormatBuildInfo returns build time and go version info
func FormatBuildInfo() (string, string) {
	build := buildTime
	goVer := goVersion
	if goVer == "" {
		goVer = runtime.Version()
	}
	return build, goVer
}

// GetVersion returns the version string
func GetVersion() string {
	return version
}
