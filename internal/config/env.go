package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port                  string
	FrontendURL           string
	SessionSecret         string
	HTTPShutdownTimeout   time.Duration
	HTTPReadHeaderTimeout time.Duration
}

func Load() (Config, error) {
	shutdownTimeout, err := Duration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	readHeaderTimeout, err := Duration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	return Config{
		Port:                  String("PORT", "3000"),
		FrontendURL:           String("FRONTEND_URL", "http://localhost:5173"),
		SessionSecret:         String("SESSION_SECRET", ""),
		HTTPShutdownTimeout:   shutdownTimeout,
		HTTPReadHeaderTimeout: readHeaderTimeout,
	}, nil
}

func String(name string, fallback string) string {
	value := normalize(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func Duration(name string, fallback time.Duration) (time.Duration, error) {
	value := normalize(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	return duration, nil
}

func normalize(value string) string {
	value = strings.TrimSpace(value)
	if len(value) < 2 {
		return value
	}
	if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
		return strings.TrimSpace(value[1 : len(value)-1])
	}
	return value
}
