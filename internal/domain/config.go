package domain

type QbitConfig struct {
	Addr      string `mapstructure:"addr"`
	Host      string `mapstructure:"host"`
	Port      uint   `mapstructure:"port"`
	Login     string `mapstructure:"login"`
	Password  string `mapstructure:"password"`
	BasicUser string `mapstructure:"basicUser"`
	BasicPass string `mapstructure:"basicPass"`
}

type ReannounceSettings struct {
	Enabled  bool `mapstructure:"enabled"`
	Attempts int  `mapstructure:"attempts"`
	Interval int  `mapstructure:"interval"`
}

type Rules struct {
	Enabled            bool `mapstructure:"enabled"`
	MaxActiveDownloads int  `mapstructure:"max_active_downloads"`
}

type AppConfig struct {
	Debug      bool               `mapstructure:"debug"`
	Qbit       QbitConfig         `mapstructure:"qbittorrent"`
	Reannounce ReannounceSettings `mapstructure:"reannounce"`
	Rules      Rules              `mapstructure:"rules"`
	Compare    []QbitConfig       `mapstructure:"compare"`
}
