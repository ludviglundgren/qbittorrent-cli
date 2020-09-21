package domain

type QbitConfig struct {
	Host     string `mapstructure:"host"`
	Port     uint   `mapstructure:"port"`
	Login    string `mapstructure:"login"`
	Password string `mapstructure:"password"`
}

type AppConfig struct {
	Debug bool       `mapstructure:"debug"`
	Qbit  QbitConfig `mapstructure:"qbittorrent"`
}
