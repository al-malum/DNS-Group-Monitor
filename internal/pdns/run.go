package pdns

import (
	"log/slog"
	"main/pkg/web"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Структура, где поля будут содержать дескрипторы метрик dns
type DnsMetricsDesc struct {
	AllServers         *prometheus.Desc
	AvailabileServers  *prometheus.Desc
	UnavailableServers *prometheus.Desc
	MaintenanceServers *prometheus.Desc
}

// глобальное определение конфигурации и ошибки чтения (если есть)
var Conf, ConfErr = GetConfig()

// Реализация интерфейса collector
// метод Describe возвращает описание(дескриптор) всех метрик собранных этим коллектором в выделенный канал
func (DnsMetrics *DnsMetricsDesc) Describe(ch chan<- *prometheus.Desc) {
	ch <- DnsMetrics.AllServers // экземпляр DnsMetrics создается в функции NewDnsMetrics()
	ch <- DnsMetrics.AvailabileServers
	ch <- DnsMetrics.UnavailableServers
	ch <- DnsMetrics.MaintenanceServers
}

// метод Collect возвращает в канал саму метрику и вызывается каждый раз при получении данных
// так же возвращается дескриптор метрики
// дескриптор, который передает Collect должен быть одним из тех, что возвращает Describe
// метрики, использующие один и тот же дескриптор, должны отличаться лейблами
func (DnsMetrics *DnsMetricsDesc) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	var resultCheckingAuth []AvailabilityGroup
	chAvailGrp := make(chan []AvailabilityGroup, 50)

	wg.Add(1) // Увеличиваем счетчик для двух горутин

	go func() {
		defer wg.Done() // Уменьшаем счетчик в конце выполнения горутины
		CheckAvailabilityDns(Conf.GroupsDNS, chAvailGrp)
	}()

	wg.Wait()
	close(chAvailGrp)

	seenMetrics := make(map[string]struct{}) // Карта для отслеживания уже отправленных метрик

	for result := range chAvailGrp {
		resultCheckingAuth = append(resultCheckingAuth, result...)
	}

	for _, item := range resultCheckingAuth {
		// Создаем уникальный ключ для метрики, используя значение label "cluster"
		key := item.GroupName

		// Проверяем, была ли уже отправлена такая метрика
		if _, exists := seenMetrics[key]; exists {
			continue // Пропускаем, если метрика уже была отправлена
		}
		// Отметим, что метрика была отправлена
		seenMetrics[key] = struct{}{}
		// Отправляем метрики в канал
		ch <- prometheus.MustNewConstMetric(
			DnsMetrics.AllServers,
			prometheus.GaugeValue,
			float64(item.AllServers),
			item.GroupName,
		)
		ch <- prometheus.MustNewConstMetric(
			DnsMetrics.AvailabileServers,
			prometheus.GaugeValue,
			float64(item.AvailabileServers),
			item.GroupName,
		)
		ch <- prometheus.MustNewConstMetric(
			DnsMetrics.UnavailableServers,
			prometheus.GaugeValue,
			float64(item.UnavailableServers),
			item.GroupName,
		)
		ch <- prometheus.MustNewConstMetric(
			DnsMetrics.MaintenanceServers,
			prometheus.GaugeValue,
			float64(item.MaintenanceServers),
			item.GroupName,
		)
	}
}

// Создание нового объекта, структуры, полем которой является дескриптор (дескрипторы) метрик
func NewDnsMetrics() *DnsMetricsDesc {
	return &DnsMetricsDesc{
		AllServers: prometheus.NewDesc(
			"all_servers", // имя метрики
			"Общее количество кластеров днс в составе большого кластера", // хелп метрики
			[]string{"server"},  // variableLabels, лейблы метрики в зависимости от входящих данных при формировании метрики в методе Collect()
			prometheus.Labels{}, // constLabels, заранее определяемые лейблы метрик этого типа (опционально)
		),
		AvailabileServers: prometheus.NewDesc(
			"available_servers", // имя метрики
			"Количество доступных кластеров днс в составе большого кластера", // хелп метрики
			[]string{"server"},  // variableLabels, лейблы метрики в зависимости от входящих данных при формировании метрики в методе Collect()
			prometheus.Labels{}, // constLabels, заранее определяемые лейблы метрик этого типа (опционально)
		),
		UnavailableServers: prometheus.NewDesc(
			"unavailable_servers", // имя метрики
			"Количество недоступных кластеров днс в составе большого кластера", // хелп метрики
			[]string{"server"},  // variableLabels, лейблы метрики в зависимости от входящих данных при формировании метрики в методе Collect()
			prometheus.Labels{}, // constLabels, заранее определяемые лейблы метрик этого типа (опционально)
		),
		MaintenanceServers: prometheus.NewDesc(
			"maintenance_servers", // имя метрики
			"Количество кластеров днс в составе большого кластера на обслуживании", // хелп метрики
			[]string{"server"},  // variableLabels, лейблы метрики в зависимости от входящих данных при формировании метрики в методе Collect()
			prometheus.Labels{}, // constLabels, заранее определяемые лейблы метрик этого типа (опционально)
		),
	}
}

func Run() error {
	if ConfErr != nil {
		return ConfErr
	}
	initLogger(Conf.LogPath, Conf.LogLevel)
	reg := prometheus.NewPedanticRegistry()
	workerDns := NewDnsMetrics()
	mtlsSett := web.MtlsSettings{
		Enabled:   Conf.MtlsExporter.Enabled,
		Key:       Conf.MtlsExporter.Key,
		Cert:      Conf.MtlsExporter.Cert,
		AllowedCN: Conf.MtlsExporter.AllowedCN,
	}
	reg.MustRegister(workerDns)
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	http.Handle("/metrics", web.AuthenticationCN(promHandler, mtlsSett))
	if Conf.MtlsExporter.Enabled {
		slog.Info("Run server with mtls.")
		RunServerWithTls(promHandler, Conf.MtlsExporter)
	} else {
		slog.Info("Run server without mtls.")
		RunServerWithousTls(promHandler)
	}
	return nil
}
