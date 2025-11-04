package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	ServerPort	int			`json:"server_port"`
	DBHost			string	`json:"db_host"`
	DBPort			int			`json:"db_port"`
	DBDatabase	string	`json:"db_database"`
	DBUsername	string	`json:"db_username"`
	DBPassword	string	`json:"db_password"`
}

func LoadConfig() (*Config, error) {
	configPath := filepath.Join("config", "config.json")
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}