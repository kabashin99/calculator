package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AddTime time.Duration
	SubTime time.Duration
	MulTime time.Duration
	DivTime time.Duration
}

func Load() *Config {
	return &Config{
		AddTime: getEnvDuration("TIME_ADDITION_MS", 1000),
		SubTime: getEnvDuration("TIME_SUBTRACTION_MS", 2000),
		MulTime: getEnvDuration("TIME_MULTIPLICATION_MS", 3000),
		DivTime: getEnvDuration("TIME_DIVISION_MS", 4000),
	}
}

func getEnvDuration(key string, defaultVal int) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return time.Duration(defaultVal) * time.Millisecond
	}
	ms, _ := strconv.Atoi(val)
	return time.Duration(ms) * time.Millisecond
}

func (c *Config) Get() *Config {
	return c
}
