package pdns

import (
	"sync"
)

// Структура, которая идентифицирует авторити кластер и содержит отчет о доступности кластеров в его составе
type AvailabilityGroup struct {
	GroupName          string
	AllServers         int8
	AvailabileServers  int8
	UnavailableServers int8
	MaintenanceServers int8
}

// Функция обработки групп DNS серверов
func processingDnsGroup(group GroupDNS) AvailabilityGroup {
	dnsClient := CreateDnsClient()
	chDns := make(chan DnsResponseData, len(group.DNSServers))
	var wg sync.WaitGroup
	availGroup := AvailabilityGroup{
		GroupName:          group.GroupName,
		AllServers:         int8(len(group.DNSServers)),
		AvailabileServers:  0,
		UnavailableServers: 0,
		MaintenanceServers: 0,
	}
	counter := 0
	for _, target := range group.DNSServers {
		counter++
		if target.Maintenance {
			availGroup.MaintenanceServers++
			continue
		}
		wg.Add(1)
		dnsReqData := CreateDnsRequestData(target.ServerID, target.IP, target.RequestedRecord, int32(target.DNSPort))
		go DnsRequest(dnsReqData, chDns, dnsClient, &wg)
	}
	wg.Wait()
	close(chDns)

	for result := range chDns {
		switch result.Availability {
		case true:
			availGroup.AvailabileServers++
		case false:
			availGroup.UnavailableServers++
		}
	}
	// go func() {
	// 	defer wg.Wait()
	// 	close(chDns)
	// }()
	return availGroup
}

// Основная функция
func CheckAvailabilityDns(dnsGroups []GroupDNS, chAvailMgcl chan []AvailabilityGroup) {
	var wgAvailAuth sync.WaitGroup
	var availList []AvailabilityGroup

	for _, group := range dnsGroups {
		wgAvailAuth.Add(1)
		go func(group GroupDNS) {

			defer wgAvailAuth.Done()
			dataAvail := processingDnsGroup(group)
			availList = append(availList, dataAvail)
		}(group)
	}
	wgAvailAuth.Wait()
	chAvailMgcl <- availList
}
