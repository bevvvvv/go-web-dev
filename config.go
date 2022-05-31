package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
)

func LoadConfig(required bool, dbEnv bool) AppConfig {
	var appConfig AppConfig

	jsonFile, err := os.Open("config/config.json")
	if err != nil {
		if required {
			fmt.Println("Unable to load configuration file.")
			panic(err)
		}
		fmt.Println("Using the default config...")
		appConfig = DefaultAppConfig()
	} else {
		decoder := json.NewDecoder(jsonFile)
		err = decoder.Decode(&appConfig)
		if err != nil {
			panic(err)
		}
		fmt.Println("Successfully loaded configuration json...")
	}

	// anyone with access can use account
	appConfig.Mailgun.APIKey = os.Getenv("API_KEY")
	if appConfig.Mailgun.APIKey == "" {
		panic(errors.New("No API key provided for mailgun client!"))
	}
	if dbEnv {
		appConfig.Database.Host = os.Getenv("DATABASE_HOST")
		appConfig.Database.Port, err = strconv.Atoi(os.Getenv("DATABASE_PORT"))
		if err != nil {
			panic(err)
		}
		appConfig.Database.User = os.Getenv("DATABASE_USER")
		appConfig.Database.Password = os.Getenv("DATABASE_PASSWORD")
		appConfig.Database.Name = os.Getenv("DATABASE_NAME")
	}
	return appConfig
}

type AppConfig struct {
	Port     int            `json:"port"`
	Env      string         `json:"env"`
	Pepper   string         `json:"pepper"`
	HMACKey  string         `json:"hmac_key"`
	Mailgun  MailgunConfig  `json:"mailgun"`
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

type MailgunConfig struct {
	APIKey       string `json:"api_key"`
	PublicAPIKey string `json:"public_api_key"`
	Domain       string `json:"domain"`
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
