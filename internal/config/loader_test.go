package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadConfig("non_existent_file.txt")

	assert.NoError(t, err)
	assert.Equal(t, defaultTimeAdditionMS, cfg.TimeAdditionMS)
	assert.Equal(t, defaultTimeSubtractionMS, cfg.TimeSubtractionMS)
	assert.Equal(t, defaultTimeMultiplicationMS, cfg.TimeMultiplicationMS)
	assert.Equal(t, defaultTimeDivisionMS, cfg.TimeDivisionMS)
	assert.Equal(t, defaultComputingPower, cfg.ComputingPower)
}

func TestLoadConfig_ValidFile(t *testing.T) {
	configContent := `TIME_ADDITION_MS=150
TIME_SUBTRACTION_MS=120
TIME_MULTIPLICATION_MS=250
TIME_DIVISION_MS=300
COMPUTING_POWER=8
`

	tmpFile := createTempConfigFile(t, configContent)
	defer os.Remove(tmpFile)

	cfg, err := LoadConfig(tmpFile)

	assert.NoError(t, err)
	assert.Equal(t, 150, cfg.TimeAdditionMS)
	assert.Equal(t, 120, cfg.TimeSubtractionMS)
	assert.Equal(t, 250, cfg.TimeMultiplicationMS)
	assert.Equal(t, 300, cfg.TimeDivisionMS)
	assert.Equal(t, 8, cfg.ComputingPower)
}

func TestLoadConfig_InvalidValues(t *testing.T) {
	configContent := `TIME_ADDITION_MS=abc
TIME_SUBTRACTION_MS=-50
TIME_MULTIPLICATION_MS=300
TIME_DIVISION_MS=xyz
COMPUTING_POWER=NaN
`

	tmpFile := createTempConfigFile(t, configContent)
	defer os.Remove(tmpFile)

	cfg, err := LoadConfig(tmpFile)

	assert.NoError(t, err)
	assert.Equal(t, defaultTimeAdditionMS, cfg.TimeAdditionMS)
	assert.Equal(t, -50, cfg.TimeSubtractionMS)
	assert.Equal(t, 300, cfg.TimeMultiplicationMS)
	assert.Equal(t, defaultTimeDivisionMS, cfg.TimeDivisionMS)
	assert.Equal(t, defaultComputingPower, cfg.ComputingPower)
}

func createTempConfigFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "config_*.txt")
	assert.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)

	tmpFile.Close()
	return tmpFile.Name()
}
