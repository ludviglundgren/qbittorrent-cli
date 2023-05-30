package config

import (
	"fmt"
	"os"

	"github.com/ludviglundgren/qbittorrent-cli/internal/domain"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	CfgFile    string
	Config     domain.AppConfig
	Qbit       domain.QbitConfig
	Compare    domain.QBCompareSettings
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

		// Search config in directories
		viper.SetConfigName(".qbt")
		viper.AddConfigPath(".") // optionally look for config in the working directory
		viper.AddConfigPath(home)
		viper.AddConfigPath("$HOME/.config/qbt") // call multiple times to add many search paths
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Could not read config file:", viper.ConfigFileUsed())
		os.Exit(1)
	}

	err := viper.Unmarshal(&Config)
	if err != nil {
		os.Exit(1)
	}
	Qbit = Config.Qbit
	Compare = Config.Compare
	Reannounce = Config.Reannounce
	Rules = Config.Rules
}
