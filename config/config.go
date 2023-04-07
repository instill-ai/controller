package config

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/redis/go-redis/v9"
)

// Config - Global variable to export
var Config AppConfig

// AppConfig defines
type AppConfig struct {
	Server           ServerConfig           `koanf:"server"`
	Etcd             EtcdConfig             `koanf:"etcd"`
	Database         DatabaseConfig         `koanf:"database"`
	Cache            CacheConfig            `koanf:"cache"`
	TritonServer     TritonServerConfig     `koanf:"tritonserver"`
	ConnectorBackend ConnectorBackendConfig `koanf:"connectorbackend"`
	ModelBackend     ModelBackendConfig     `koanf:"modelbackend"`
	PipelineBackend  PipelineBackendConfig  `koanf:"pipelinebackend"`
	MgmtBackend      MgmtBackendConfig      `koanf:"mgmtbackend"`
}

// ServerConfig defines HTTP server configurations
type ServerConfig struct {
	Port  int `koanf:"privateport"`
	HTTPS struct {
		Cert string `koanf:"cert"`
		Key  string `koanf:"key"`
	}
	CORSOrigins  []string      `koanf:"corsorigins"`
	Edition      string        `koanf:"edition"`
	Debug        bool          `koanf:"debug"`
	LoopInterval time.Duration `koanf:"loopinterval"`
	Timeout      time.Duration `koanf:"timeout"`
}

// EtcdConfig
type EtcdConfig struct {
	Host    string        `koanf:"host"`
	Port    string        `koanf:"port"`
	Timeout time.Duration `koanf:"timeout"`
}

// DatabaseConfig related to database
type DatabaseConfig struct {
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Name     string `koanf:"name"`
	Version  uint   `koanf:"version"`
	TimeZone string `koanf:"timezone"`
	Pool     struct {
		IdleConnections int           `koanf:"idleconnections"`
		MaxConnections  int           `koanf:"maxconnections"`
		ConnLifeTime    time.Duration `koanf:"connlifetime"`
	}
}

// CacheConfig related to Redis
type CacheConfig struct {
	Redis struct {
		RedisOptions redis.Options `koanf:"redisoptions"`
	}
}

// TritonServerConfig related to Triton server
type TritonServerConfig struct {
	Host       string `koanf:"host"`
	GrpcURI    string `koanf:"grpcuri"`
	ModelStore string `koanf:"modelstore"`
}

// ConnectorBackendConfig related to connector-backend
type ConnectorBackendConfig struct {
	Host        string `koanf:"host"`
	PublicPort  int    `koanf:"publicport"`
	PrivatePort int    `koanf:"privateport"`
	HTTPS       struct {
		Cert string `koanf:"cert"`
		Key  string `koanf:"key"`
	}
}

// ModelBackendConfig related to model-backend
type ModelBackendConfig struct {
	Host        string `koanf:"host"`
	PublicPort  int    `koanf:"publicport"`
	PrivatePort int    `koanf:"privateport"`
	HTTPS       struct {
		Cert string `koanf:"cert"`
		Key  string `koanf:"key"`
	}
}

// PipelineBackendConfig related to pipeline-backend
type PipelineBackendConfig struct {
	Host        string `koanf:"host"`
	PublicPort  int    `koanf:"publicport"`
	PrivatePort int    `koanf:"privateport"`
	HTTPS       struct {
		Cert string `koanf:"cert"`
		Key  string `koanf:"key"`
	}
}

// MgmtBackendConfig related to mgmt-backend
type MgmtBackendConfig struct {
	Host        string `koanf:"host"`
	PublicPort  int    `koanf:"publicport"`
	PrivatePort int    `koanf:"privateport"`
	HTTPS       struct {
		Cert string `koanf:"cert"`
		Key  string `koanf:"key"`
	}
}

// Init - Assign global config to decoded config struct
func Init() error {
	k := koanf.New(".")
	parser := yaml.Parser()

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fileRelativePath := fs.String("file", "config/config.yaml", "configuration file")
	flag.Parse()

	if err := k.Load(file.Provider(*fileRelativePath), parser); err != nil {
		log.Fatal(err.Error())
	}

	if err := k.Load(env.ProviderWithValue("CFG_", ".", func(s string, v string) (string, interface{}) {
		key := strings.Replace(strings.ToLower(strings.TrimPrefix(s, "CFG_")), "_", ".", -1)
		if strings.Contains(v, ",") {
			return key, strings.Split(strings.TrimSpace(v), ",")
		}
		return key, v
	}), nil); err != nil {
		return err
	}

	if err := k.Unmarshal("", &Config); err != nil {
		return err
	}

	return ValidateConfig(&Config)
}

// ValidateConfig is for custom validation rules for the configuration
func ValidateConfig(cfg *AppConfig) error {
	return nil
}
