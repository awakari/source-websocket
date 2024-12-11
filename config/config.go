package config

import (
	"github.com/kelseyhightower/envconfig"
	"time"
)

type Config struct {
	Api ApiConfig
	Db  DbConfig
	Log struct {
		Level int `envconfig:"LOG_LEVEL" default:"-4" required:"true"`
	}
	Replica ReplicaConfig
}

type ApiConfig struct {
	Port   uint16 `envconfig:"API_PORT" default:"50051" required:"true"`
	Writer struct {
		Backoff time.Duration `envconfig:"API_WRITER_BACKOFF" default:"10s" required:"true"`
		Uri     string        `envconfig:"API_WRITER_URI" default:"http://pub:8080/v1" required:"true"`
	}
	Token struct {
		Internal string `envconfig:"API_TOKEN_INTERNAL" required:"true"`
	}
	UserAgent string `envconfig:"API_USER_AGENT" default:"Awakari" required:"true"`
	GroupId   string `envconfig:"API_GROUP_ID" default:"default" required:"true"`
	Events    EventsConfig
}

type EventsConfig struct {
	Source string `envconfig:"API_EVENTS_SOURCE" default:"https://awakari.com/pub.html?srcType=ws" required:"true"`
	Type   string `envconfig:"API_EVENTS_TYPE" required:"true" default:"com_awakari_websocket_v1"`
}

type DbConfig struct {
	Uri      string `envconfig:"DB_URI" default:"mongodb://localhost:27017/?retryWrites=true&w=majority" required:"true"`
	Name     string `envconfig:"DB_NAME" default:"sources" required:"true"`
	UserName string `envconfig:"DB_USERNAME" default:""`
	Password string `envconfig:"DB_PASSWORD" default:""`
	Table    struct {
		Name      string        `envconfig:"DB_TABLE_NAME" default:"websocket" required:"true"`
		Retention time.Duration `envconfig:"DB_TABLE_RETENTION" default:"2160h" required:"true"`
		Shard     bool          `envconfig:"DB_TABLE_SHARD" default:"true"`
	}
	Tls struct {
		Enabled  bool `envconfig:"DB_TLS_ENABLED" default:"false" required:"true"`
		Insecure bool `envconfig:"DB_TLS_INSECURE" default:"false" required:"true"`
	}
}

type ReplicaConfig struct {
	Count uint32 `envconfig:"REPLICA_COUNT" required:"true"`
	Name  string `envconfig:"REPLICA_NAME" required:"true"`
}

type WriterCacheConfig struct {
	Size uint32        `envconfig:"API_WRITER_CACHE_SIZE" default:"100" required:"true"`
	Ttl  time.Duration `envconfig:"API_WRITER_CACHE_TTL" default:"24h" required:"true"`
}

func NewConfigFromEnv() (cfg Config, err error) {
	err = envconfig.Process("", &cfg)
	return
}
