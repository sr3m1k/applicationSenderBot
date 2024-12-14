package config

import (
	"github.com/pelletier/go-toml"
	"log"
	"os"
)

type Config struct {
	Bot struct {
		Token string `toml:"token"`
	} `toml:"bot"`
	Access struct {
		Users  []int64 `toml:"users"`
		Admins []int64 `toml:"admins"`
	} `toml:"access"`
	Database struct {
		Path string `toml:"path"`
	} `toml:"database"`
}

func LoadConfig(configPath string) (Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return Config{}, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("Не удалось закрыть файл конфигурации: %v", err)
		}
	}()

	var config Config
	decoder := toml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
