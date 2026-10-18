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

	S3Endpoint  string `env:"S3_ENDPOINT"`
	S3AccessKey string `env:"S3_ACCESS_KEY"`
	S3SecretKey string `env:"S3_SECRET_KEY"`
	S3Bucket    string `env:"S3_BUCKET,default=raas-media-bucket"`
	S3Region    string `env:"AWS_REGION,default=us-east-1"`
	// S3PublicURL is the browser-reachable base URL stored in media records; defaults to endpoint/bucket.
	S3PublicURL string `env:"S3_PUBLIC_URL"`

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
