package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadConfig(required bool) AppConfig {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		if required {
			fmt.Println("Unable to load configuration file.")
			panic(err)
		}
		fmt.Println("Using the default config...")
		return DefaultAppConfig()
	}
	var appConfig AppConfig
	decoder := json.NewDecoder(jsonFile)
	err = decoder.Decode(&appConfig)
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully loaded configuration json...")
	return appConfig
}

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (pgConfig PostgresConfig) Dialect() string {
	return "postgres"
}

func (pgConfig PostgresConfig) ConnectionString() string {
	if pgConfig.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
			pgConfig.Host, pgConfig.Port, pgConfig.User, pgConfig.Name)
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		pgConfig.Host, pgConfig.Port, pgConfig.User, pgConfig.Password, pgConfig.Name)
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "host.docker.internal",
		Port:     5432,
		User:     "postgres",
		Password: "secretpass",
		Name:     "fakeoku",
	}
}

type AppConfig struct {
	Port     int            `json:"port"`
	Env      string         `json:"env"`
	Pepper   string         `json:"pepper"`
	HMACKey  string         `json:"hmac_key"`
	Database PostgresConfig `json:"database"`
}

func (appConfig *AppConfig) IsProd() bool {
	return appConfig.Env == "prod"
}

func DefaultAppConfig() AppConfig {
	return AppConfig{
		Port:     3000,
		Env:      "dev",
		Pepper:   "dev-pepper",
		HMACKey:  "dev-hmac-key",
		Database: DefaultPostgresConfig(),
	}
}
