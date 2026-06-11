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
	Model       ModelConfig       `mapstructure:"model"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type InfluxDBConfig struct {
	Addr           string `mapstructure:"addr"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	Database       string `mapstructure:"database"`
	Precision      string `mapstructure:"precision"`
	WriteQueueSize int    `mapstructure:"write_queue_size"`
	WriteMaxRetries int   `mapstructure:"write_max_retries"`
}

type LoRaConfig struct {
	DataEndpoint string `mapstructure:"data_endpoint"`
	DeviceCount  int    `mapstructure:"device_count"`
	UDPAddr      string `mapstructure:"udp_addr"`
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

type ModelConfig struct {
	LstmPath       string `mapstructure:"lstm_path"`
	LstmInputSize  int    `mapstructure:"lstm_input_size"`
	LstmHiddenSize int    `mapstructure:"lstm_hidden_size"`
	LstmOutputSize int    `mapstructure:"lstm_output_size"`
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
