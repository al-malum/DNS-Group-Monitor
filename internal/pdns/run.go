package pdns

import (
	"log/slog"
	"main/pkg/web"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// DnsMetricsDesc содержит дескрипторы метрик для мониторинга состояния DNS серверов в группе
type DnsMetricsDesc struct {
	AllServers         *prometheus.Desc // Дескриптор метрики для общего количества DNS серверов в группе
	AvailabileServers  *prometheus.Desc // Дескриптор метрики для доступных DNS серверов
	UnavailableServers *prometheus.Desc // Дескриптор метрики для недоступных DNS серверов
	MaintenanceServers *prometheus.Desc // Дескриптор метрики для серверов, находящихся на обслуживании
}

// глобальные переменные для конфигурации и ошибок при чтении конфигурации
var Conf, ConfErr = GetConfig()

// Describe реализует интерфейс prometheus.Collector, описывая метрики, которые будет собирать данный коллектор
// В канал ch передаются дескрипторы всех метрик, собранных этим коллектором
func (DnsMetrics *DnsMetricsDesc) Describe(ch chan<- *prometheus.Desc) {
	// Логирование описания метрик
	slog.Debug("Describing DNS metrics.")
	// Описание всех метрик DNS
	ch <- DnsMetrics.AllServers
	ch <- DnsMetrics.AvailabileServers
	ch <- DnsMetrics.UnavailableServers
	ch <- DnsMetrics.MaintenanceServers
}

// Collect реализует интерфейс prometheus.Collector, собирая метрики для мониторинга
// В канал ch передаются сами метрики для Prometheus
// Каждый раз, когда вызывается Collect, обновляются значения метрик
func (DnsMetrics *DnsMetricsDesc) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	var resultCheckingAuth []AvailabilityGroup
	chAvailGrp := make(chan []AvailabilityGroup, 50) // Канал для передачи результатов проверки доступности

	wg.Add(1) // Увеличиваем счетчик горутин

	// Логируем начало сбора метрик
	slog.Debug("Starting collection of DNS metrics.")
	go func() {
		defer wg.Done() // Уменьшаем счетчик горутин по завершению
		// Выполняем проверку доступности DNS серверов для каждой группы
		slog.Debug("Checking DNS server availability.")
		CheckAvailabilityDns(Conf.GroupsDNS, chAvailGrp)
	}()

	// Ожидаем завершения горутины
	wg.Wait()
	close(chAvailGrp)

	seenMetrics := make(map[string]struct{}) // Карта для отслеживания уже отправленных метрик

	// Обрабатываем результаты проверки доступности DNS серверов
	for result := range chAvailGrp {
		resultCheckingAuth = append(resultCheckingAuth, result...)
	}
	// Логируем количество групп, для которых будут отправлены метрики
	slog.Debug("Sending metrics for groups.", slog.Int("num_groups", len(resultCheckingAuth)))
	// Отправляем метрики для каждой группы
	for _, item := range resultCheckingAuth {
		// Создаем уникальный ключ для метрики на основе имени группы
		key := item.GroupName

		// Проверяем, была ли уже отправлена метрика для данной группы
		if _, exists := seenMetrics[key]; exists {
			continue // Пропускаем, если метрика уже была отправлена
		}
		// Отметим, что метрика была отправлена
		seenMetrics[key] = struct{}{}

		// Отправляем метрики в канал Prometheus
		slog.Debug("Sending metric for group.", slog.String("group", item.GroupName))
		ch <- prometheus.MustNewConstMetric(
			DnsMetrics.AllServers, // Метрика общего количества серверов
			prometheus.GaugeValue,
			float64(item.AllServers), // Значение метрики
			item.GroupName,           // Лейбл, идентифицирующий группу серверов
		)
		ch <- prometheus.MustNewConstMetric(
			DnsMetrics.AvailabileServers, // Метрика доступных серверов
			prometheus.GaugeValue,
			float64(item.AvailabileServers),
			item.GroupName,
		)
		ch <- prometheus.MustNewConstMetric(
			DnsMetrics.UnavailableServers, // Метрика недоступных серверов
			prometheus.GaugeValue,
			float64(item.UnavailableServers),
			item.GroupName,
		)
		ch <- prometheus.MustNewConstMetric(
			DnsMetrics.MaintenanceServers, // Метрика серверов на обслуживании
			prometheus.GaugeValue,
			float64(item.MaintenanceServers),
			item.GroupName,
		)
	}
}

// NewDnsMetrics создает новый объект DnsMetricsDesc с дескрипторами для метрик DNS серверов
// Каждая метрика будет собираться с лейблом, соответствующим группе серверов
func NewDnsMetrics() *DnsMetricsDesc {
	return &DnsMetricsDesc{
		AllServers: prometheus.NewDesc(
			"all_servers", // Имя метрики для общего количества серверов
			"Total number of DNS servers in the group", // Описание метрики
			[]string{"group"},                          // Лейблы метрики: идентификатор группы серверов
			prometheus.Labels{},                        // Нет предустановленных лейблов
		),
		AvailabileServers: prometheus.NewDesc(
			"available_servers",                            // Имя метрики для доступных серверов
			"Number of available DNS servers in the group", // Описание метрики
			[]string{"group"},                              // Лейблы метрики: идентификатор группы серверов
			prometheus.Labels{},                            // Нет предустановленных лейблов
		),
		UnavailableServers: prometheus.NewDesc(
			"unavailable_servers",                            // Имя метрики для недоступных серверов
			"Number of unavailable DNS servers in the group", // Описание метрики
			[]string{"group"},                                // Лейблы метрики: идентификатор группы серверов
			prometheus.Labels{},                              // Нет предустановленных лейблов
		),
		MaintenanceServers: prometheus.NewDesc(
			"maintenance_servers", // Имя метрики для серверов на обслуживании
			"Number of DNS servers in the group under maintenance", // Описание метрики
			[]string{"group"},   // Лейблы метрики: идентификатор группы серверов
			prometheus.Labels{}, // Нет предустановленных лейблов
		),
	}
}

// Run инициализирует сервер и запускает сбор метрик для Prometheus
// В зависимости от конфигурации может быть включен mTLS для безопасного соединения
func Run() error {
	if ConfErr != nil {
		// Логируем ошибку при чтении конфигурации
		slog.Error("Error reading configuration", "error", ConfErr)
		return ConfErr // Если ошибка при чтении конфигурации, возвращаем ошибку
	}

	// Инициализация логгера с заданными параметрами
	initLogger(Conf.LogPath, Conf.LogLevel, Conf.LogToFile, Conf.LogToSyslog)

	// Логируем успешное чтение конфигурации
	slog.Info("Configuration loaded successfully.")

	// Регистрируем коллектор метрик для Prometheus
	reg := prometheus.NewPedanticRegistry()
	workerDns := NewDnsMetrics()

	// Настройки для mTLS (если включен)
	mtlsSett := web.MtlsSettings{
		Enabled:   Conf.MtlsExporter.Enabled,
		Key:       Conf.MtlsExporter.Key,
		Cert:      Conf.MtlsExporter.Cert,
		AllowedCN: Conf.MtlsExporter.AllowedCN,
	}

	// Регистрируем наш коллектор метрик в Prometheus
	reg.MustRegister(workerDns)

	// Обрабатываем запросы к меткам с использованием mTLS или без него
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	http.Handle("/metrics", web.AuthenticationCN(promHandler, mtlsSett))

	if Conf.MtlsExporter.Enabled {
		// Запускаем сервер с поддержкой mTLS
		slog.Info("Run server with mtls.")
		RunServerWithTls(promHandler, Conf.MtlsExporter)
	} else {
		// Запускаем сервер без mTLS
		slog.Info("Run server without mtls.")
		RunServerWithousTls(promHandler)
	}

	return nil
}
