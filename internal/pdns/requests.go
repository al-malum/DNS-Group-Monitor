package pdns

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// DnsResponseData хранит результаты выполнения DNS запроса:
// - ID сервера,
// - время отклика,
// - сообщение с ответом от сервера,
// - доступность сервера (успешно ли выполнен запрос).
type DnsResponseData struct {
	ServerID       string        // Идентификатор сервера
	TimeToResponse time.Duration // Время отклика от DNS сервера
	Msg            *dns.Msg      // Сообщение с ответом DNS сервера
	Availability   bool          // Указывает, был ли сервер доступен (запрос успешен)
}

// DnsRequestData содержит данные, необходимые для выполнения DNS запроса:
// - ID сервера,
// - адрес и порт сервера,
// - полностью квалифицированное доменное имя (FQDN).
type DnsRequestData struct {
	ServerID string // Идентификатор сервера
	Address  string // IP адрес или хостнейм DNS сервера
	Fqdn     string // Полностью квалифицированное доменное имя для запроса
	Port     int32  // Порт DNS сервера
}

// CreateDnsRequestData создает и возвращает структуру DnsRequestData с необходимыми данными для DNS запроса
// Аргументы:
// - serverID: идентификатор DNS сервера,
// - address: адрес DNS сервера (IP или хостнейм),
// - record: FQDN для DNS запроса,
// - dnsPort: порт для DNS запроса.
func CreateDnsRequestData(serverID, address, record string, dnsPort int32) DnsRequestData {
	return DnsRequestData{
		ServerID: serverID,
		Address:  address,
		Fqdn:     record,
		Port:     dnsPort,
	}
}

// CreateDnsClient создает и настраивает новый DNS клиент с нужными тайм-аутами для операций записи и чтения.
// Возвращает указатель на сконфигурированный DNS клиент.
func CreateDnsClient() *dns.Client {
	var dnsClient dns.Client
	// Настройка Dialer для соединений с DNS сервером
	dnsClient.Dialer = &net.Dialer{
		Timeout: 1 * time.Second, // Тайм-аут соединения с DNS сервером
	}
	// Устанавливаем тайм-ауты для операций чтения и записи
	dnsClient.ReadTimeout = 2 * time.Second
	dnsClient.WriteTimeout = 2 * time.Second
	return &dnsClient
}

// DnsRequest выполняет DNS запрос к указанному серверу и передает результат в канал chDns.
// Для выполнения используется DNS клиент, переданный в качестве аргумента.
// После выполнения запроса горутина завершает работу (defer wg.Done()).
func DnsRequest(drd DnsRequestData, chDns chan DnsResponseData, dnsClient *dns.Client, wg *sync.WaitGroup) {
	defer wg.Done() // Обеспечиваем, что горутина завершится при выходе из функции
	var (
		msg        dns.Msg // Сообщение для запроса
		checkAvail bool    // Флаг доступности DNS сервера
	)
	// Формируем запрос DNS на основе FQDN
	fqdn := dns.Fqdn(drd.Fqdn)
	msg.SetQuestion(fqdn, dns.TypeA) // Устанавливаем тип запроса (A-запись)
	// Выполняем запрос к DNS серверу по указанному адресу и порту
	resp, ttr, err := dnsClient.Exchange(&msg, fmt.Sprintf("%s:%d", drd.Address, drd.Port))
	if err != nil {
		// В случае ошибки считаем сервер недоступным
		checkAvail = false
	} else {
		// Если запрос успешен, считаем сервер доступным
		checkAvail = true
	}
	// Формируем структуру с результатами запроса
	responseDns := DnsResponseData{
		ServerID:       drd.ServerID,                      // Идентификатор сервера
		Availability:   checkAvail,                        // Доступность сервера
		TimeToResponse: time.Duration(ttr.Milliseconds()), // Время отклика сервера (в миллисекундах)
		Msg:            resp,                              // Ответ от DNS сервера
	}
	// Отправляем результат в канал для дальнейшей обработки
	chDns <- responseDns
}
