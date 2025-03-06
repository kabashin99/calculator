package config

import (
	"bufio"
	"fmt"
	"log"
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
		log.Printf("Config file not found, using default values: %v\n", cfg)
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
			if v, err := strconv.Atoi(value); err == nil {
				cfg.TimeAdditionMS = v
			}
		case "TIME_SUBTRACTION_MS":
			if v, err := strconv.Atoi(value); err == nil {
				cfg.TimeSubtractionMS = v
			}
		case "TIME_MULTIPLICATION_MS":
			if v, err := strconv.Atoi(value); err == nil {
				cfg.TimeMultiplicationMS = v
			}
		case "TIME_DIVISION_MS":
			if v, err := strconv.Atoi(value); err == nil {
				cfg.TimeDivisionMS = v
			}
		case "COMPUTING_POWER":
			if v, err := strconv.Atoi(value); err == nil {
				cfg.ComputingPower = v
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	log.Printf("Config loaded: %v\n", cfg)
	return cfg, nil
}
