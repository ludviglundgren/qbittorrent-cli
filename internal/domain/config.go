package domain

type QbitConfig struct {
	Addr      string `mapstructure:"addr"`
	Host      string `mapstructure:"host"`
	Port      uint   `mapstructure:"port"`
	Login     string `mapstructure:"login"`
	Password  string `mapstructure:"password"`
	APIKey    string `mapstructure:"apikey"`
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

type AddConfig struct {
	Sequential     bool `mapstructure:"sequential"`
	FirstLastPiece bool `mapstructure:"first_last_piece"`
}

type AppConfig struct {
	Debug      bool               `mapstructure:"debug"`
	Qbit       QbitConfig         `mapstructure:"qbittorrent"`
	Reannounce ReannounceSettings `mapstructure:"reannounce"`
	Rules      Rules              `mapstructure:"rules"`
	Add        AddConfig          `mapstructure:"add"`
	Compare    []QbitConfig       `mapstructure:"compare"`
}
