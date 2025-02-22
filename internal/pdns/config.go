package pdns

import (
	"encoding/json"
	"flag"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
)

// Config - основная структура конфигурации, которая содержит параметры для работы приложения.
// Включает в себя:
// - путь к файлу логов,
// - уровень логирования,
// - настройки для mTLS экспорта,
// - группы DNS серверов.
type Config struct {
	LogPath      string     `json:"logPath"`      // Путь к файлу логов
	LogLevel     string     `json:"logLevel"`     // Уровень логирования
	LogToFile    bool       `json:"logToFile"`    // Логирование в файл
	LogToSyslog  bool       `json:"logToSyslog"`  // Логирование в syslog
	MtlsExporter MtlsConfig `json:"mtlsExporter"` // Конфигурация mTLS
	GroupsDNS    []GroupDNS `json:"groupsDns"`    // Список групп DNS серверов
}

// MtlsConfig - структура для конфигурации mTLS (mutual TLS).
// Содержит параметры для включения mTLS и настройки безопасности.
type MtlsConfig struct {
	Enabled     bool     `json:"enabled"`     // Флаг включения mTLS
	Key         string   `json:"key"`         // Путь к приватному ключу
	Cert        string   `json:"cert"`        // Путь к сертификату
	AllowedCN   []string `json:"allowedCN"`   // Список допустимых Common Name (CN)
	Description string   `json:"description"` // Описание настроек
}

// GroupDNS - структура, представляющая группу DNS серверов.
// Содержит:
// - имя группы,
// - список DNS серверов в этой группе.
type GroupDNS struct {
	GroupName  string      `json:"groupName"`  // Имя группы DNS серверов
	DNSServers []DNSTarget `json:"dnsServers"` // Список DNS серверов в группе
}

// DNSTarget - структура, содержащая информацию о конкретном DNS сервере.
// Включает в себя:
// - уникальный идентификатор сервера,
// - IP адрес,
// - порт,
// - запрашиваемую запись,
// - состояние обслуживания,
// - описание сервера.
type DNSTarget struct {
	ServerID        string `json:"serverID"`        // Идентификатор сервера
	IP              string `json:"IP"`              // IP адрес DNS сервера
	DNSPort         int    `json:"dnsPort"`         // Порт DNS сервера
	RequestedRecord string `json:"requestedRecord"` // Запрашиваемая DNS запись (например, A-запись)
	Maintenance     bool   `json:"maintenance"`     // Флаг, указывающий на состояние обслуживания
	Description     string `json:"description"`     // Описание DNS сервера
}

// GetConfig читает конфигурацию из JSON файла, указанного через флаг `-c`.
// Возвращает объект конфигурации Config или ошибку в случае неудачи.
// Также выполняет валидацию конфигурации с помощью библиотеки validator.
func GetConfig() (*Config, error) {
	var path string
	// Чтение флага с путем к файлу конфигурации
	flag.StringVar(&path, "c", "/etc/dns-group-monitor/config.json", "path to config file")
	flag.Parse()

	// Логирование пути к конфигурационному файлу
	slog.Debug("Reading configuration file", slog.String("configFilePath", path))

	// Чтение содержимого конфигурационного файла
	plan, errRead := os.ReadFile(path)
	if errRead != nil {
		// Логируем ошибку чтения файла и возвращаем ошибку
		slog.Error("Error reading configuration file", slog.String("configFilePath", path), slog.String("error", errRead.Error()))
		return nil, errRead
	}

	var Conf Config
	// Разбираем содержимое файла в структуру Config
	err := json.Unmarshal(plan, &Conf)

	// Инициализация валидатора и проверка соответствия структуры Config
	validate := validator.New()
	if err := validate.Struct(Conf); err != nil {
		// Логируем ошибки валидации полей структуры
		errs := err.(validator.ValidationErrors)
		for _, fieldErr := range errs {
			slog.Error("Validation error", slog.String("field", fieldErr.Namespace()), slog.String("tag", fieldErr.ActualTag()), slog.String("param", fieldErr.Param()))
		}
		return nil, err // Возвращаем ошибку валидации
	}

	if err != nil {
		return &Conf, err
	}

	// Возвращаем структуру с конфигурацией
	return &Conf, nil
}

// ContainBool проверяет, содержится ли значение `key` в списке `listing` типа []bool.
// Возвращает true, если значение найдено, и false в противном случае.
func ContainBool(listing []bool, key bool) bool {
	for _, value := range listing {
		if key == value {
			return true
		}
	}
	return false
}

// ContainString проверяет, содержится ли строка `key` в списке `listing` типа []string.
// Возвращает true, если строка найдена, и false в противном случае.
func ContainString(listing []string, key string) bool {
	for _, value := range listing {
		if key == value {
			return true
		}
	}
	return false
}
