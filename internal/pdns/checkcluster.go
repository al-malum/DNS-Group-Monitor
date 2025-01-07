package pdns

import (
	"log/slog"
	"sync"
)

// AvailabilityGroup - структура, представляющая собой отчет о доступности группы DNS серверов
type AvailabilityGroup struct {
	GroupName          string // Имя группы серверов DNS
	AllServers         int8   // Общее количество серверов в группе
	AvailabileServers  int8   // Количество доступных серверов
	UnavailableServers int8   // Количество недоступных серверов
	MaintenanceServers int8   // Количество серверов на обслуживании
}

// processingDnsGroup - функция для обработки конкретной группы DNS серверов
func processingDnsGroup(group GroupDNS) AvailabilityGroup {
	dnsClient := CreateDnsClient()                             // Создаем DNS клиент с заданными настройками (тайм-ауты и т.д.)
	chDns := make(chan DnsResponseData, len(group.DNSServers)) // Канал для получения данных о каждом сервере
	var wg sync.WaitGroup                                      // Ожидание завершения всех горутин
	// Инициализируем структуру для хранения результатов обработки группы
	availGroup := AvailabilityGroup{
		GroupName:          group.GroupName,
		AllServers:         int8(len(group.DNSServers)), // Общее количество серверов в группе
		AvailabileServers:  0,                           // Изначально доступных серверов нет
		UnavailableServers: 0,                           // Изначально недоступных серверов нет
		MaintenanceServers: 0,                           // Изначально серверы на обслуживании не учитываются
	}
	counter := 0 // Счетчик, который отслеживает количество обработанных серверов

	// Логируем начало обработки группы
	slog.Info("Processing DNS group", slog.String("groupName", group.GroupName), slog.Int("serversCount", len(group.DNSServers)))

	// Проходим по каждому серверу из группы и проверяем его состояние
	for _, target := range group.DNSServers {
		counter++
		if target.Maintenance { // Если сервер находится на обслуживании, увеличиваем счетчик и пропускаем его
			availGroup.MaintenanceServers++
			slog.Debug("Server is under maintenance", slog.String("serverID", target.ServerID), slog.String("serverIP", target.IP))
			continue
		}
		wg.Add(1) // Увеличиваем счетчик горутин для каждого запроса
		// Создаем данные для DNS запроса
		dnsReqData := CreateDnsRequestData(target.ServerID, target.IP, target.RequestedRecord, int32(target.DNSPort))
		// Запускаем горутину для отправки DNS запроса асинхронно
		go DnsRequest(dnsReqData, chDns, dnsClient, &wg)
	}

	// Ожидаем завершения всех горутин
	wg.Wait()
	close(chDns) // Закрываем канал после завершения всех горутин

	// Обрабатываем результаты запросов, полученные через канал
	for result := range chDns {
		// Логируем результаты доступности каждого сервера
		if result.Availability {
			slog.Debug("Server is available", slog.String("serverID", result.ServerID), slog.String("serverIP", result.ServerID), slog.Duration("responseTime", result.TimeToResponse))
		} else {
			slog.Debug("Server is unavailable", slog.String("serverID", result.ServerID), slog.String("serverIP", result.ServerID))
		}

		// Если сервер доступен, увеличиваем счетчик доступных серверов, иначе недоступных
		switch result.Availability {
		case true:
			availGroup.AvailabileServers++
		case false:
			availGroup.UnavailableServers++
		}
	}

	// Логируем итоговые результаты для группы
	slog.Info("Finished processing DNS group",
		slog.String("groupName", availGroup.GroupName),
		slog.Int("allServers", int(availGroup.AllServers)),
		slog.Int("availableServers", int(availGroup.AvailabileServers)),
		slog.Int("unavailableServers", int(availGroup.UnavailableServers)),
		slog.Int("maintenanceServers", int(availGroup.MaintenanceServers)),
	)

	// Возвращаем итоговый результат по группе
	return availGroup
}

// CheckAvailabilityDns - основная функция для проверки доступности всех DNS серверов во всех группах
func CheckAvailabilityDns(dnsGroups []GroupDNS, chAvailMgcl chan []AvailabilityGroup) {
	var wgAvailAuth sync.WaitGroup    // Ожидание завершения всех горутин по обработке групп
	var availList []AvailabilityGroup // Список для хранения результатов по всем группам

	// Логируем начало проверки доступности всех групп
	slog.Info("Starting to check availability for all DNS groups")

	// Обрабатываем каждую группу DNS серверов
	for _, group := range dnsGroups {
		wgAvailAuth.Add(1) // Увеличиваем счетчик горутин для каждой группы

		go func(group GroupDNS) {
			defer wgAvailAuth.Done() // Уменьшаем счетчик горутин по завершению

			// Обрабатываем группу и получаем результаты
			dataAvail := processingDnsGroup(group)
			// Добавляем результат в общий список
			availList = append(availList, dataAvail)
		}(group) // Передаем группу в горутину
	}

	// Ожидаем завершения всех горутин
	wgAvailAuth.Wait()

	// Логируем завершение проверки доступности всех групп
	slog.Info("Finished checking availability for all DNS groups")

	// Отправляем собранные результаты в канал
	chAvailMgcl <- availList
}
