package config

import "time"

type Config struct {
	AppEnv string       `mapstructure:"APP_ENV"`
	Grpc   GrpcConfig   `mapstructure:",squash"`
	Logger LoggerConfig `mapstructure:",squash"`
	Cafe   CafeConfig   `mapstructure:",squash"`
}

type LoggerConfig struct {
	LogLevel     string `mapstructure:"LOG_LEVEL" validate:"required,oneof=debug info warn error"`
	LogFormatter string `mapstructure:"LOG_FORMATTER" validate:"required,oneof=json console"`
}

type GrpcConfig struct {
	Port int `mapstructure:"GRPC_PORT" validate:"required"`
}

type CafeConfig struct {
	BrewTimeout       time.Duration `mapstructure:"BREW_TIMEOUT" validate:"required"`
	OrderDurationsCap uint32        `mapstructure:"ORDER_DURATIONS_CAP" validate:"required"`
}
