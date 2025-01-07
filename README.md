# DNS Group Monitor

**DNS Group Monitor** — это инструмент для мониторинга доступности групп DNS серверов, включая как авторитативные, так и рекурсивные серверы. Проект позволяет объединить несколько DNS серверов в логические группы (например, серверы в разных дата-центрах) и отслеживать их доступность, состояние (доступен/недоступен/на обслуживании). Также реализована поддержка безопасного доступа к метрикам через mTLS с проверкой клиентского CN.

**DNS Group Monitor** is a tool for monitoring the availability of groups of DNS servers, including both authoritative and recursive servers. The project allows you to group several DNS servers into logical units (e.g., servers in different data centers) and monitor their availability and status (available/unavailable/under maintenance). It also supports secure access to metrics via mTLS with client CN verification.

## Основные особенности / Key Features

- Мониторинг доступности DNS серверов с логическим разделением на различные группы (условно, цоды).  
  - Monitoring the availability of DNS servers with logical grouping into different units (e.g., data centers).
  
- Поддержка авторитативных и рекурсивных серверов.  
  - Support for both authoritative and recursive servers.

- Отчетность о состоянии групп серверов (сколько серверов доступно, недоступно, на обслуживании).  
  - Reporting the status of server groups (how many servers are available, unavailable, or under maintenance).

- Настройка безопасного доступа к метрикам с помощью mTLS.  
  - Configuration of secure access to metrics using mTLS.

- Экспорт метрик в формате Prometheus.  
  - Metrics export in Prometheus format.

---

## Пример конфигурации / Example Configuration

Пример конфигурационного файла `config.json` (в рабочем файле комментарии не допускаются):  
Example configuration file `config.json` (comments are not allowed in the actual configuration file):

```json
{
    "logPath": "/etc/gdns-exporter/dnsexporter.log",   // Путь к файлу логов, где будут сохраняться логи работы приложения.  
    "logLevel": "INFO",                                 // Уровень логирования. Может быть: "DEBUG", "INFO", "WARN", "ERROR".  
    "mtlsExporter": {                                   // Настройки для mTLS (двусторонняя TLS аутентификация).  
        "enabled": false,                               // Включить ли mTLS для экспорта метрик. Если true, будет использоваться TLS с проверкой клиентского сертификата.  
        "key": "/etc/gdns-exporter/tls/key.pem",        // Путь к приватному ключу сервера для mTLS.  
        "cert": "/etc/gdns-exporter/tls/cert.pem",      // Путь к публичному сертификату сервера для mTLS.  
        "allowedCN": ["localhost2", "localhost1"],      // Список разрешённых значений CN (Common Name) для клиентских сертификатов. Если mTLS включен, то только клиенты с указанным CN смогут подключиться.  
        "description": "mtls for the exporter page"     // Описание модуля mTLS для экспорта метрик.  
    },  
    "groupsDns": [                                      // Массив групп DNS серверов. Каждая группа содержит несколько серверов DNS.  
        {
            "groupName": "NY Data Center",               // Название группы DNS серверов (например, для группы серверов в определённом дата-центре).  
            "dnsServers": [                              // Список DNS серверов в этой группе.  
                {
                    "serverID": "pdns-auth-1.1",         // Уникальный идентификатор сервера в группе.  
                    "IP": "8.8.8.8",                    // IP адрес DNS сервера.  
                    "dnsPort": 53,                      // Порт DNS сервера (обычно 53).  
                    "requestedRecord": "yandex.ru",     // Запрашиваемая запись (например, для проверки доступности этого DNS сервера).  
                    "maintenance": false,               // Флаг, указывающий, находится ли сервер в обслуживании. Если true, то сервер не проверяется на доступность.  
                    "description": ""                   // Дополнительное описание сервера.  
                },
                {
                    "serverID": "pdns-auth-1.2",         // Уникальный идентификатор для другого сервера.  
                    "IP": "8.8.8.8",                    // IP адрес второго DNS сервера.  
                    "dnsPort": 53,                      // Порт DNS сервера.  
                    "requestedRecord": "chatgpt.com",   // Запрашиваемая запись для второго сервера.  
                    "maintenance": false,               // Флаг, указывающий, находится ли сервер в обслуживании.  
                    "description": ""                   // Дополнительное описание второго сервера.  
                }
            ]
        },
        {
            "groupName": "MSA Data Center",               // Название второй группы DNS серверов (например, для другой локации или дата-центра).  
            "dnsServers": [                              // Список DNS серверов для этой группы.  
                {
                    "serverID": "pdns-auth-2.1",         // Уникальный идентификатор первого сервера в группе.  
                    "IP": "8.8.4.4",                    // IP адрес первого DNS сервера.  
                    "dnsPort": 53,                      // Порт DNS сервера.  
                    "requestedRecord": "example.com",   // Запрашиваемая запись для первого сервера.  
                    "maintenance": false,               // Флаг обслуживания для первого сервера.  
                    "description": ""                   // Описание для первого сервера.  
                },
                {
                    "serverID": "pdns-auth-2.2",         // Уникальный идентификатор второго сервера.  
                    "IP": "8.8.4.4",                    // IP адрес второго DNS сервера.  
                    "dnsPort": 53,                      // Порт DNS сервера.  
                    "requestedRecord": "powerdns.com",  // Запрашиваемая запись для второго сервера.  
                    "maintenance": false,               // Флаг обслуживания для второго сервера.  
                    "description": ""                   // Описание второго сервера.  
                }
            ]
        }
    ]
}
```

---

## Установка / Installation

### С помощью Docker / Using Docker

1. Склонируйте репозиторий:  
    Clone the repository:
    ```bash
    git clone https://github.com/yourusername/dns-group-monitor.git
    cd dns-group-monitor
    ```

2. Постройте Docker образ:  
    Build the Docker image:
    ```bash
    docker build -t dns-group-monitor .
    ```

3. Запустите контейнер:  
    Run the container:
    ```bash
    docker run -d -p 8080:8080 dns-group-monitor
    ```

### Сборка с помощью Go / Building with Go

1. Убедитесь, что у вас установлен Go (версии 1.16 и выше).  
    Make sure Go is installed (version 1.16 or later).

2. Склонируйте репозиторий:  
    Clone the repository:
    ```bash
    git clone https://github.com/yourusername/dns-group-monitor.git
    cd dns-group-monitor
    ```

3. Соберите проект:  
    Build the project:
    ```bash
    go build -ldflags "-X main.desiredPathPid=/etc/dns-monitor/dns-monitor.pid" cmd/pdns/main.go
    ```

    Замените `/etc/dns-monitor/dns-monitor.pid` на актуальный путь к PID файлу в вашей системе.  
    Replace `/etc/dns-monitor/dns-monitor.pid` with the actual path to the PID file on your system.

4. Запустите приложение:  
    Run the application:
    ```bash
    ./dns-group-monitor
    ```

---

## Зависимости / Dependencies

Для работы проекта используются следующие зависимости:  
The project depends on the following libraries:

- [github.com/go-playground/validator/v10](https://github.com/go-playground/validator)
- [github.com/miekg/dns](https://github.com/miekg/dns)
- [github.com/prometheus/client_golang/prometheus](https://github.com/prometheus/client_golang/prometheus)
- [github.com/prometheus/client_golang/prometheus/promhttp](https://github.com/prometheus/client_golang/prometheus/promhttp)

---

## Использование / Usage

1. Заполните конфигурационный файл `config.json` (пример конфигурации приведен выше).  
   Fill in the `config.json` configuration file (an example configuration is provided above).
   
2. Соберите проект, используя один из методов (Docker или Go).  
   Build the project using one of the methods (Docker or Go).
   
3. Запустите приложение, которое будет отдавать метрики в формате Prometheus.  
   Run the application, which will expose metrics in the Prometheus format.
   
4. Подключите приложение к Prometheus для мониторинга.  
   Connect the application to Prometheus for monitoring.

---

## Лицензия / License

Проект распространяется под лицензией MIT.  
This project is licensed under the MIT License.

