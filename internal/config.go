package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Service struct {
		Port int `yaml:"port"`
	} `yaml:"service"`
	Data struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Database string `yaml:"database"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"data"`
}

func ReadConfig(path string) (*Config, error) {
	// Read config
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config %w", err)
	}

	var config Config

	if err = yaml.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml %w", err)
	}

	log.Println("Successfully read config file.")

	return &config, nil
}

func (config *Config) ConnectionString() string {
	dbHost := net.JoinHostPort(config.Data.Host, strconv.Itoa(config.Data.Port))

	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		config.Data.User, config.Data.Password, dbHost, config.Data.Database)
}
