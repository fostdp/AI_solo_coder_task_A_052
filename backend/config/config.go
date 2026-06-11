package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	InfluxDB    InfluxDBConfig    `mapstructure:"influxdb"`
	LoRa        LoRaConfig        `mapstructure:"lora"`
	Alert       AlertConfig       `mapstructure:"alert"`
	Fumigation  FumigationConfig  `mapstructure:"fumigation"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type InfluxDBConfig struct {
	Addr      string `mapstructure:"addr"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Database  string `mapstructure:"database"`
	Precision string `mapstructure:"precision"`
}

type LoRaConfig struct {
	DataEndpoint string `mapstructure:"data_endpoint"`
	DeviceCount  int    `mapstructure:"device_count"`
}

type AlertConfig struct {
	AcousticEventThreshold float64 `mapstructure:"acoustic_event_threshold"`
	MoistureThreshold      float64 `mapstructure:"moisture_threshold"`
	WechatWebhookURL       string  `mapstructure:"wechat_webhook_url"`
	SmsAPIURL              string  `mapstructure:"sms_api_url"`
	SmsAPIKey              string  `mapstructure:"sms_api_key"`
}

type FumigationConfig struct {
	DefaultReleaseRate float64 `mapstructure:"default_release_rate"`
	WindSpeed          float64 `mapstructure:"wind_speed"`
	WindDirection      float64 `mapstructure:"wind_direction"`
	StabilityClass     string  `mapstructure:"stability_class"`
}

var AppConfig *Config

func LoadConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
