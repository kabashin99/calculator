package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultTimeAdditionMS       = 100
	defaultTimeSubtractionMS    = 100
	defaultTimeMultiplicationMS = 200
	defaultTimeDivisionMS       = 200
	defaultComputingPower       = 4
)

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		TimeAdditionMS:       defaultTimeAdditionMS,
		TimeSubtractionMS:    defaultTimeSubtractionMS,
		TimeMultiplicationMS: defaultTimeMultiplicationMS,
		TimeDivisionMS:       defaultTimeDivisionMS,
		ComputingPower:       defaultComputingPower,
	}

	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Config file not found, using default values: %v\n", cfg)
		return cfg, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "TIME_ADDITION_MS":
			cfg.TimeAdditionMS, _ = strconv.Atoi(value)
		case "TIME_SUBTRACTION_MS":
			cfg.TimeSubtractionMS, _ = strconv.Atoi(value)
		case "TIME_MULTIPLICATION_MS":
			cfg.TimeMultiplicationMS, _ = strconv.Atoi(value)
		case "TIME_DIVISION_MS":
			cfg.TimeDivisionMS, _ = strconv.Atoi(value)
		case "COMPUTING_POWER":
			cfg.ComputingPower, _ = strconv.Atoi(value)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	fmt.Printf("Config loaded: %v\n", cfg)
	return cfg, nil
}
