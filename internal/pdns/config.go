package pdns

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
)

// Config - общая структура конфигурации
type Config struct {
	LogPath      string     `json:"logPath"`
	LogLevel     string     `json:"logLevel"`
	MtlsExporter MtlsConfig `json:"mtlsExporter"`
	GroupsDNS    []GroupDNS `json:"groupsDns"`
}

// MtlsConfig - конфигурация mtls
type MtlsConfig struct {
	Enabled     bool     `json:"enabled"`
	Key         string   `json:"key"`
	Cert        string   `json:"cert"`
	AllowedCN   []string `json:"allowedCN"`
	Description string   `json:"description"`
}

// GroupDNS - группа DNS-серверов
type GroupDNS struct {
	GroupName  string      `json:"groupName"`
	DNSServers []DNSTarget `json:"dnsServers"`
}

// DNSTarget - информация о конкретном DNS-сервере
type DNSTarget struct {
	ServerID        string `json:"serverID"`
	IP              string `json:"IP"`
	DNSPort         int    `json:"dnsPort"`
	RequestedRecord string `json:"requestedRecord"`
	Maintenance     bool   `json:"maintenance"`
	Description     string `json:"description"`
}

// Функция для чтения конфигурационного файла
func GetConfig() (*Config, error) {
	var path string
	flag.StringVar(&path, "c", "/etc/ddidnser/config.json", "path to config file")
	flag.Parse()
	plan, errRead := os.ReadFile(path)
	if errRead != nil {
		slog.Error(errRead.Error())
		return nil, errRead
	}
	var Conf Config
	err := json.Unmarshal(plan, &Conf)

	validate := validator.New()
	if err := validate.Struct(Conf); err != nil {
		errs := err.(validator.ValidationErrors)
		for _, fieldErr := range errs {
			slog.Error(fmt.Sprintf("field %s %s %s\n", fieldErr.Namespace(), fieldErr.ActualTag(), fieldErr.Param()))
		}
		return nil, err
	}

	if err != nil {
		return &Conf, err
	}
	return &Conf, nil
}

func ContainBool(listing []bool, key bool) bool {
	for _, value := range listing {
		if key == value {
			return true
		}
	}
	return false
}

func ContainString(listing []string, key string) bool {
	for _, value := range listing {
		if key == value {
			return true
		}
	}
	return false
}
