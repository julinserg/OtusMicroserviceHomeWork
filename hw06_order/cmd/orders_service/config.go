package main

import "github.com/BurntSushi/toml"

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger LoggerConf
	PSQL   PSQLConfig
	HTTP   HTTPConfig
}

type LoggerConf struct {
	Level string
	// TODO
}

type PSQLConfig struct {
	DSN string
}

type HTTPConfig struct {
	Host string
	Port string
}

func (c *Config) Read(fpath string) error {
	_, err := toml.DecodeFile(fpath, c)
	return err
}

func NewConfig() Config {
	return Config{}
}
