package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SavesDir  string `yaml:"saves_dir"`
	SteamDir  string `yaml:"steam_dir"`
	Character string `yaml:"character"`
}

func LoadConfig() (*Config, error) {
	// Load config from file
	config := &Config{}
	bytes, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func SaveConfig(config *Config) error {
	// Save config to file
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	err = os.WriteFile("config.yaml", bytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func NewConfig() *Config {
	return &Config{
		SavesDir:  "",
		SteamDir:  "",
		Character: "",
	}
}

func IsExist() bool {
	_, err := os.Stat("config.yaml")
	return err == nil
}
