package config

import "github.com/spf13/viper"

type Config struct {
	ServerPort string `mapstructure:"SERVER_PORT"`
	MongoURI   string `mapstructure:"MONGO_URI"`

	// SMTP 설정
	SMTPHost     string `mapstructure:"SMTP_HOST"`
	SMTPPort     int    `mapstructure:"SMTP_PORT"`
	SMTPUsername string `mapstructure:"SMTP_USERNAME"`
	SMTPPassword string `mapstructure:"SMTP_PASSWORD"`
	SMTPFrom     string `mapstructure:"SMTP_FROM"`

	// RabbitMQ 설정
	RabbitMQURL     string `mapstructure:"RABBITMQ_URL"`
	RabbitMQQueue   string `mapstructure:"RABBITMQ_QUEUE"`
	RabbitMQRetries int    `mapstructure:"RABBITMQ_RETRIES"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// 기본값 설정
	viper.SetDefault("SERVER_PORT", "8003")
	viper.SetDefault("SMTP_PORT", 587)
	viper.SetDefault("RABBITMQ_QUEUE", "email_queue")
	viper.SetDefault("RABBITMQ_RETRIES", 3)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
