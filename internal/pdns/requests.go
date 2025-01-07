package pdns

import (
	"fmt"
	"net"

	"sync"
	"time"

	"github.com/miekg/dns"
)

// Структура, где поля будут содержать результаты запроса
type DnsResponseData struct {
	ServerID       string
	TimeToResponse time.Duration
	Msg            *dns.Msg
	Availability   bool
}

// Структура, необходимая для днс запроса
type DnsRequestData struct {
	ServerID string
	Address  string
	Fqdn     string
	Port     int32
}

// Функия для создание структуры с данными для запроса dns
func CreateDnsRequestData(serverID, address, record string, dnsPort int32) DnsRequestData {
	return DnsRequestData{
		ServerID: serverID,
		Address:  address,
		Fqdn:     record,
		Port:     dnsPort,
	}
}

// Функция для создание днс клиента
func CreateDnsClient() *dns.Client {
	var dnsClient dns.Client
	// Устанавливаем настройки для Dialer
	dnsClient.Dialer = &net.Dialer{
		Timeout: 1 * time.Second, // Устанавливаем тайм-аут в 1 секунду
	}
	// Устанавливаем тайм-ауты для чтения и записи
	dnsClient.ReadTimeout = 2 * time.Second
	dnsClient.WriteTimeout = 2 * time.Second
	return &dnsClient
}

// Функция по выполнению днс запросов
func DnsRequest(drd DnsRequestData, chDns chan DnsResponseData, dnsClient *dns.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	var (
		msg        dns.Msg
		checkAvail bool
	)
	fmt.Println(drd.Address, drd.Fqdn, drd.Port, drd.ServerID)
	fqdn := dns.Fqdn(drd.Fqdn)
	msg.SetQuestion(fqdn, dns.TypeA)
	resp, ttr, err := dnsClient.Exchange(&msg, fmt.Sprintf("%s:%d", drd.Address, drd.Port)) // выполнение запроса
	fmt.Println(resp)
	if err != nil {
		checkAvail = false
	} else {
		checkAvail = true
	}
	// время ответа возвращается в миллисекундах, 300 - порог
	responseDns := DnsResponseData{
		ServerID:       drd.ServerID,
		Availability:   checkAvail,
		TimeToResponse: time.Duration(ttr.Milliseconds()),
		Msg:            resp,
	}
	chDns <- responseDns
}
