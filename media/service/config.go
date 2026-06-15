package service

import (
	"github.com/Netflix/go-env"
)

// Config represents all application configuration loaded from the environment.
type Config struct {
	MongoURI     string   `env:"MONGO_URI,default=mongodb://localhost:27017"`
	MongoDBName  string   `env:"MONGO_DB_NAME,default=media_db"`
	KafkaBrokers []string `env:"KAFKA_BROKERS,default=localhost:9092"`
	Port         string   `env:"PORT,default=8080"`

	R2AccountID       string `env:"R2_ACCOUNT_ID"`
	R2AccessKeyID     string `env:"R2_ACCESS_KEY_ID"`
	R2SecretAccessKey string `env:"R2_SECRET_ACCESS_KEY"`
	R2BucketName      string `env:"R2_BUCKET_NAME,default=raas"`
	R2PublicURL       string `env:"R2_PUBLIC_URL"`

	Extras env.EnvSet
}

// LoadConfig parses environment variables into the Config struct using go-env.
func LoadConfig() (*Config, error) {
	var cfg Config
	es, err := env.UnmarshalFromEnviron(&cfg)
	if err != nil {
		return nil, err
	}
	cfg.Extras = es

	return &cfg, nil
}
