package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadConfig_Success(t *testing.T) {
	content := `TIME_ADDITION_MS=150
TIME_SUBTRACTION_MS=200
TIME_MULTIPLICATION_MS=250
TIME_DIVISION_MS=300
COMPUTING_POWER=8
JWT_SECRET_KEY=some-secret-key
`
	tmpFile, err := os.CreateTemp("", "config_test_*.env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatal(err)
	}

	err = tmpFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, 150, cfg.TimeAdditionMS)
	assert.Equal(t, 200, cfg.TimeSubtractionMS)
	assert.Equal(t, 250, cfg.TimeMultiplicationMS)
	assert.Equal(t, 300, cfg.TimeDivisionMS)
	assert.Equal(t, 8, cfg.ComputingPower)
	assert.Equal(t, "some-secret-key", cfg.JwtSecretKey)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadConfig("non_existent_file.env")
	assert.NoError(t, err)

	assert.Equal(t, defaultTimeAdditionMS, cfg.TimeAdditionMS)
	assert.Equal(t, defaultTimeSubtractionMS, cfg.TimeSubtractionMS)
	assert.Equal(t, defaultTimeMultiplicationMS, cfg.TimeMultiplicationMS)
	assert.Equal(t, defaultTimeDivisionMS, cfg.TimeDivisionMS)
	assert.Equal(t, defaultComputingPower, cfg.ComputingPower)
	assert.Equal(t, defaultJwtSecretKey, cfg.JwtSecretKey)
}

func TestLoadConfig_InvalidData(t *testing.T) {
	content := `TIME_ADDITION_MS=abc
TIME_SUBTRACTION_MS=200
TIME_MULTIPLICATION_MS=xyz
TIME_DIVISION_MS=300
COMPUTING_POWER=invalid
JWT_SECRET_KEY=
`
	tmpFile, err := os.CreateTemp("", "config_invalid_test_*.env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatal(err)
	}

	err = tmpFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmpFile.Name())
	assert.NoError(t, err)

	assert.Equal(t, defaultTimeAdditionMS, cfg.TimeAdditionMS)
	assert.Equal(t, 200, cfg.TimeSubtractionMS)
	assert.Equal(t, defaultTimeMultiplicationMS, cfg.TimeMultiplicationMS)
	assert.Equal(t, 300, cfg.TimeDivisionMS)
	assert.Equal(t, defaultComputingPower, cfg.ComputingPower)
	assert.Equal(t, defaultJwtSecretKey, cfg.JwtSecretKey)
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config_empty_test_*.env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	err = tmpFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmpFile.Name())
	assert.NoError(t, err)

	assert.Equal(t, defaultTimeAdditionMS, cfg.TimeAdditionMS)
	assert.Equal(t, defaultTimeSubtractionMS, cfg.TimeSubtractionMS)
	assert.Equal(t, defaultTimeMultiplicationMS, cfg.TimeMultiplicationMS)
	assert.Equal(t, defaultTimeDivisionMS, cfg.TimeDivisionMS)
	assert.Equal(t, defaultComputingPower, cfg.ComputingPower)
	assert.Equal(t, defaultJwtSecretKey, cfg.JwtSecretKey)
}
