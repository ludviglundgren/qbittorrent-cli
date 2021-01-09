package domain

type QbitConfig struct {
	Host     string `mapstructure:"host"`
	Port     uint   `mapstructure:"port"`
	Login    string `mapstructure:"login"`
	Password string `mapstructure:"password"`
}

type ReannounceSettings struct {
	Enabled  bool `mapstructure:"enabled"`
	Attempts int  `mapstructure:"attempts"`
	Interval int  `mapstructure:"interval"`
}

type AppConfig struct {
	Debug      bool               `mapstructure:"debug"`
	Qbit       QbitConfig         `mapstructure:"qbittorrent"`
	Reannounce ReannounceSettings `mapstructure:"reannounce"`
}
