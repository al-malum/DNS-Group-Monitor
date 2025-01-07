Вот пример README файла для вашего проекта, основываясь на предоставленной информации:

---

# DNS Group Monitor

**DNS Group Monitor** — это инструмент для мониторинга доступности групп DNS серверов, включая как авторитативные, так и рекурсивные серверы. Проект позволяет объединить несколько DNS серверов в логические группы (например, серверы в разных дата-центрах) и отслеживать их доступность, состояние (доступен/недоступен/на обслуживании). Также реализована поддержка безопасного доступа к метрикам через mTLS с проверкой клиентского CN.

## Основные особенности
- Мониторинг доступности DNS серверов с логическим разделением на различные группы (условно, цоды).
- Поддержка авторитативных и рекурсивных серверов.
- Отчетность о состоянии групп серверов (сколько серверов доступно, недоступно, на обслуживании).
- Настройка безопасного доступа к метрикам с помощью mTLS.
- Экспорт метрик в формате Prometheus.

## Пример конфигурации

Пример конфигурационного файла `config.json`:

```json
{
    "logPath": "/etc/gdns-exporter/dnsexporter.log",
    "logLevel": "INFO",
    "mtlsExporter": {
        "enabled": false,
        "key": "/etc/gdns-exporter/tls/key.pem",
        "cert": "/etc/gdns-exporter/tls/cert.pem",
        "allowedCN": ["localhost2", "localhost1"],
        "description": "mtls for the exporter page"
    },
    "groupsDns": [
        {
            "groupName": "NY Data Center",
            "dnsServers": [
                {
                    "serverID": "pdns-auth-1.1",
                    "IP": "8.8.8.8",
                    "dnsPort": 53,
                    "requestedRecord": "yandex.ru",
                    "maintenance": false,
                    "description": ""
                },
                {
                    "serverID": "pdns-auth-1.2",
                    "IP": "8.8.8.8",
                    "dnsPort": 53,
                    "requestedRecord": "chatgpt.com",
                    "maintenance": false,
                    "description": ""
                }
            ]
        },
        {
            "groupName": "MSA Data Center",
            "dnsServers": [
                {
                    "serverID": "pdns-auth-2.1",
                    "IP": "8.8.4.4",
                    "dnsPort": 53,
                    "requestedRecord": "example.com",
                    "maintenance": false,
                    "description": ""
                },
                {
                    "serverID": "pdns-auth-2.2",
                    "IP": "8.8.4.4",
                    "dnsPort": 53,
                    "requestedRecord": "powerdns.com",
                    "maintenance": false,
                    "description": ""
                }
            ]
        }
    ]
}
```

## Установка

### С помощью Docker

1. Склонируйте репозиторий:
    ```bash
    git clone https://github.com/yourusername/dns-group-monitor.git
    cd dns-group-monitor
    ```

2. Постройте Docker образ:
    ```bash
    docker build -t dns-group-monitor .
    ```

3. Запустите контейнер:
    ```bash
    docker run -d -p 8080:8080 dns-group-monitor
    ```

### Сборка с помощью Go

1. Убедитесь, что у вас установлен Go (версии 1.16 и выше).

2. Склонируйте репозиторий:
    ```bash
    git clone https://github.com/yourusername/dns-group-monitor.git
    cd dns-group-monitor
    ```

3. Соберите проект:
    ```bash
    go build -ldflags "-X main.desiredPathPid=/home/user/dns-group-monitor/gdns.pid" cmd/pdns/main.go
    ```

    Замените `/home/user/dns-group-monitor/gdns.pid` на актуальный путь к PID файлу в вашей системе.

4. Запустите приложение:
    ```bash
    ./dns-group-monitor
    ```

## Зависимости

Для работы проекта используются следующие зависимости:

- [github.com/go-playground/validator/v10](https://github.com/go-playground/validator)
- [github.com/miekg/dns](https://github.com/miekg/dns)
- [github.com/prometheus/client_golang/prometheus](https://github.com/prometheus/client_golang/prometheus)
- [github.com/prometheus/client_golang/prometheus/promhttp](https://github.com/prometheus/client_golang/prometheus/promhttp)

## Использование

1. Заполните конфигурационный файл `config.json` (пример конфигурации приведен выше).
2. Соберите проект, используя один из методов (Docker или Go).
3. Запустите приложение, которое будет отдавать метрики в формате Prometheus.
4. Подключите приложение к Prometheus для мониторинга.

## Лицензия

Проект распространяется под лицензией MIT.

