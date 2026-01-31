package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// GetEnv возвращает значение переменной окружения или defaultValue если не задана.
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvRequired возвращает значение переменной окружения или паникует если не задана.
func GetEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("required environment variable not set: " + key)
	}
	return value
}

// GetEnvInt возвращает int значение переменной окружения или defaultValue.
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvInt64 возвращает int64 значение переменной окружения или defaultValue.
func GetEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvBool возвращает bool значение переменной окружения или defaultValue.
// Принимает: "true", "1", "yes" как true; "false", "0", "no" как false.
func GetEnvBool(key string, defaultValue bool) bool {
	value := strings.ToLower(os.Getenv(key))
	switch value {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return defaultValue
	}
}

// GetEnvDuration возвращает time.Duration значение переменной окружения или defaultValue.
// Формат: "5s", "1m", "2h30m" и т.д.
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// GetEnvSlice возвращает slice строк из переменной окружения, разделённых separator.
func GetEnvSlice(key string, separator string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, separator)
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}
