package config

type Config struct {
	App        App        `toml:"app"`
	Connection Connection `toml:"connection"`
}

type App struct {
	Mode           string `toml:"mode"`
	Debug          bool   `toml:"debug"`
	Name           string `toml:"name"`
	URL            string `toml:"url"`
	Port           int    `toml:"port"`
	JWTSecret      string `toml:"jwt_secret"`
	Prod           bool   `toml:"prod"`
	MediaDir       string `toml:"media_dir"`
	AdminOrigin    string `toml:"admin_origin"`
	CSRFKey        string `toml:"csrf_key"`
	MaxUploadBytes int64  `toml:"max_upload_bytes"`
}

// Connection
type Connection struct {
	Postgresql Postgresql `toml:"postgresql"`
}

type Postgresql struct {
	DSN                    string `toml:"dsn"`
	MaxOpenConnections     int    `toml:"max_open_connections"`
	MaxIdleConnections     int    `toml:"max_idle_connections"`
	MaxLifetimeConnections int    `toml:"max_lifetime_connections"`
	EncryptionKey          string `toml:"encryption_key"`
}
