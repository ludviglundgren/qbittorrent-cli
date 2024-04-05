package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ludviglundgren/qbittorrent-cli/internal/domain"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	CfgFile    string
	Config     domain.AppConfig
	Qbit       domain.QbitConfig
	Compare    []domain.QbitConfig
	Reannounce domain.ReannounceSettings
	Rules      domain.Rules
)

// InitConfig initialize config
func InitConfig() {
	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println("Could not read home dir:", err)
			os.Exit(1)
		}

		viper.SetConfigName(".qbt")
		// Search config in directories
		// call multiple times to add many search paths
		viper.AddConfigPath(".") // optionally look for config in the working directory
		viper.AddConfigPath(home)
		viper.AddConfigPath(filepath.Join(home, ".config", "qbt")) // windows path
		viper.AddConfigPath("$HOME/.config/qbt")
	}

	if err := viper.ReadInConfig(); err != nil {
		var ferr *viper.ConfigFileNotFoundError
		if errors.As(err, &ferr) {
			fmt.Printf("config file not found: err %q\n", ferr)
		} else {
			fmt.Printf("could not read config: err %q\n", err)
		}
		os.Exit(1)
	}

	if err := viper.Unmarshal(&Config); err != nil {
		os.Exit(1)
	}

	Qbit = Config.Qbit
	Compare = Config.Compare
	Reannounce = Config.Reannounce
	Rules = Config.Rules
}
