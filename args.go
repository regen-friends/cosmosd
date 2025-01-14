package main

import (
	"net/url"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const (
	rootName    = "upgrade_manager"
	genesisDir  = "genesis"
	upgradesDir = "upgrades"
	currentLink = "current"
)

// Config is the information passed in to control the daemon
type Config struct {
	Home                  string
	Name                  string
	AllowDownloadBinaries bool
	RestartAfterUpgrade   bool
}

// Root returns the root directory where all info lives
func (cfg *Config) Root() string {
	return filepath.Join(cfg.Home, rootName)
}

// GenesisBin is the path to the genesis binary - must be in place to start manager
func (cfg *Config) GenesisBin() string {
	return filepath.Join(cfg.Root(), genesisDir, "bin", cfg.Name)
}

// UpgradeBin is the path to the binary for the named upgrade
func (cfg *Config) UpgradeBin(upgradeName string) string {
	return filepath.Join(cfg.UpgradeDir(upgradeName), "bin", cfg.Name)
}

// UpgradeDir is the directory named upgrade
func (cfg *Config) UpgradeDir(upgradeName string) string {
	safeName := url.PathEscape(upgradeName)
	return filepath.Join(cfg.Root(), upgradesDir, safeName)
}

// CurrentBin is the path to the currently selected binary (genesis if no link is set)
// This will resolve the symlink to the underlying directory to make it easier to debug
func (cfg *Config) CurrentBin() string {
	cur := filepath.Join(cfg.Root(), currentLink)
	// if nothing here, fallback to genesis
	info, err := os.Lstat(cur)
	if err != nil {
		return cfg.GenesisBin()
	}
	// if it is there, ensure it is a symlink
	if info.Mode()&os.ModeSymlink == 0 {
		return cfg.GenesisBin()
	}

	// resolve it
	dest, err := os.Readlink(cur)
	if err != nil {
		return cfg.GenesisBin()
	}

	// and return the binary
	return filepath.Join(dest, "bin", cfg.Name)
}

// GetConfigFromEnv will read the environmental variables into a config
// and then validate it is reasonable
func GetConfigFromEnv() (*Config, error) {
	cfg := &Config{
		Home: os.Getenv("DAEMON_HOME"),
		Name: os.Getenv("DAEMON_NAME"),
	}
	if os.Getenv("DAEMON_ALLOW_DOWNLOAD_BINARIES") == "on" {
		cfg.AllowDownloadBinaries = true
	}
	if os.Getenv("DAEMON_RESTART_AFTER_UPGRADE") == "on" {
		cfg.RestartAfterUpgrade = true
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate returns an error if this config is invalid.
// it enforces Home/upgrade_manager is a valid directory and exists,
// and that Name is set
func (cfg *Config) validate() error {
	if cfg.Name == "" {
		return errors.New("DAEMON_NAME is not set")
	}
	if cfg.Home == "" {
		return errors.New("DAEMON_HOME is not set")
	}

	if !filepath.IsAbs(cfg.Home) {
		return errors.New("DAEMON_HOME must be an absolute path")
	}

	// ensure the root directory exists
	info, err := os.Stat(cfg.Root())
	if err != nil {
		return errors.Wrap(err, "cannot stat home dir")
	}
	if !info.IsDir() {
		return errors.Errorf("%s is not a directory", info.Name())
	}

	return nil
}
