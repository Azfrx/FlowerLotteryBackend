package config

type Config struct {
	Server   Server   `mapstructure:"server"`
	Database Database `mapstructure:"database"`
	JWT      JWT      `mapstructure:"jwt"`
	Log      Log      `mapstructure:"log"`
	Storage  Storage  `mapstructure:"storage"`
}
type Server struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}
type Database struct {
	DSN                    string `mapstructure:"dsn"`
	MaxIdleConns           int    `mapstructure:"max_idle_conns"`
	MaxOpenConns           int    `mapstructure:"max_open_conns"`
	ConnMaxLifetimeMinutes int    `mapstructure:"conn_max_lifetime_minutes"`
}
type JWT struct {
	Issuer              string `mapstructure:"issuer"`
	Secret              string `mapstructure:"secret"`
	AccessExpireMinutes int    `mapstructure:"access_expire_minutes"`
	RefreshExpireHours  int    `mapstructure:"refresh_expire_hours"`
}
type Storage struct {
	UploadDir string `mapstructure:"upload_dir"`
}
type Log struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}
