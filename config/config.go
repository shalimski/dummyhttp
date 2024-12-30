package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultListenPort = ":8080"

type Config struct {
	Mode    string        `yaml:"mode"`
	Server  ServerConfig  `yaml:"server"`
	Handler HandlerConfig `yaml:"handler"`
}

type ServerConfig struct {
	Listen  string        `yaml:"listen"`
	Timeout time.Duration `yaml:"timeout,omitempty"`
}

type HandlerConfig struct {
	Message string `yaml:"message"`
	// Дополнительные поля для разных режимов
	StaticResponses map[string]string `yaml:"static_responses,omitempty"`
	OpenAPISpec     string            `yaml:"openapi_spec,omitempty"`
}

func LoadConfig() (*Config, error) {
	var (
		configPath = flag.String("config", "", "path to config file")
		mode       = flag.String("mode", "dummy", "server mode (dummy, static, openapi)")
		listen     = flag.String("listen", defaultListenPort, "address to listen on")
		timeout    = flag.Int("timeout", 5000, "server timeout in milliseconds")
		message    = flag.String("message", "hello, world", "message to return (dummy mode)")
		help       = flag.Bool("help", false, "show help")
	)

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Дефолтная конфигурация
	config := &Config{
		Mode: "dummy",
		Server: ServerConfig{
			Listen:  defaultListenPort,
			Timeout: 5000,
		},
		Handler: HandlerConfig{
			Message: "hello, world",
		},
	}

	// Если указан файл конфигурации, загружаем из него
	if *configPath != "" {
		data, err := os.ReadFile(*configPath)
		if err != nil {
			return nil, fmt.Errorf("reading config file: %w", err)
		}

		if err = yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("parsing YAML config: %w", err)
		}
	}

	if mode != nil {
		config.Mode = *mode
	}
	if listen != nil {
		config.Server.Listen = *listen
	}

	if timeout != nil {
		config.Server.Timeout = time.Duration(*timeout) * time.Millisecond
	}

	if message != nil {
		config.Handler.Message = *message
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

func validateConfig(config *Config) error {
	switch config.Mode {
	case "dummy":
		// for dummy mode, default message is used
	case "static":
		if len(config.Handler.StaticResponses) == 0 {
			return errors.New("static mode requires static_responses configuration")
		}
	case "openapi":
		if config.Handler.OpenAPISpec == "" {
			return errors.New("openapi mode requires openapi_spec configuration")
		}
	default:
		return fmt.Errorf("unknown mode: %s", config.Mode)
	}

	if config.Server.Listen == "" {
		return errors.New("listen address is required")
	}

	if config.Server.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}

	return nil
}
