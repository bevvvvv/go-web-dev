package main

import "fmt"

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
	Port int
	Env  string
}

func (appConfig *AppConfig) IsProd() bool {
	return appConfig.Env == "prod"
}

func DefaultAppConfig() AppConfig {
	return AppConfig{
		Port: 3000,
		Env:  "dev",
	}
}
