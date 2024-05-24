package config

type Config struct {
	PocketConsumerKey string
}

var instance *Config

func InitConfig(consumerKey string) {
	instance = &Config{
		PocketConsumerKey: consumerKey,
	}
}

func GetConfig() Config {
	return *instance
}
