package pdns

import (
	"crypto/tls"    // Пакет для работы с TLS (Transport Layer Security)
	"crypto/x509"   // Пакет для работы с сертификатами X.509
	"encoding/json" // Пакет для работы с JSON
	"log"           // Пакет для логирования
	"net/http"      // Пакет для создания HTTP серверов
	"os"            // Пакет для работы с операционной системой
)

// AuthenticationCN - middleware для проверки мTLS аутентификации с использованием Common Name (CN)
// Этот middleware проверяет, что CN в клиентском сертификате соответствует одному из разрешенных CN,
// указанных в конфигурации. Если мTLS отключен или CN разрешен, запрос передается следующему обработчику.
func AuthenticationCN(next http.Handler, mtlsSetting MtlsConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что запрос был отправлен через mTLS (проверка наличия верифицированной цепочки сертификатов)
		if r.TLS != nil && len(r.TLS.VerifiedChains) > 0 && len(r.TLS.VerifiedChains[0]) > 0 {
			var commonName = r.TLS.VerifiedChains[0][0].Subject.CommonName // Извлекаем Common Name (CN) из сертификата клиента
			// Если CN есть в списке разрешенных CN в конфиге, то аутентификация успешна
			if ContainString(mtlsSetting.AllowedCN, commonName) {
				log.Printf("Authentication successful. CN: %s, Remote address: %s", commonName, r.RemoteAddr)
				next.ServeHTTP(w, r) // Передаем запрос следующему обработчику
			} else if len(mtlsSetting.AllowedCN) == 0 || !mtlsSetting.Enabled {
				// Если мTLS не включен или список разрешенных CN пуст, разрешаем запрос
				log.Printf("mTLS disabled. Request without mTLS successful. Remote address: %s", r.RemoteAddr)
				next.ServeHTTP(w, r)
			} else {
				// Если CN не разрешен, возвращаем ошибку 403 (Forbidden)
				log.Printf("Authentication failed - incorrect CN: %s, Remote address: %s", commonName, r.RemoteAddr)
				w.WriteHeader(http.StatusForbidden)
				w.Header().Set("Content-Type", "application/json")
				response := make(map[string]string)
				response["message"] = "Incorrect CN of the certificate"
				jsonResponse, _ := json.Marshal(response) // Преобразуем ответ в JSON
				w.Write(jsonResponse)
			}
		} else {
			// Если mTLS не используется, разрешаем запрос
			log.Printf("mTLS disabled. Request without mTLS successful. Remote address: %s", r.RemoteAddr)
			next.ServeHTTP(w, r)
		}
	})
}

// RunServerWithTls - запускает HTTPS сервер с поддержкой mTLS
// Настроен для проверки клиентских сертификатов, используя файл сертификата и ключа
// на порту 9100. Функция ожидает подключения и обрабатывает их через TLS.
func RunServerWithTls(handler http.Handler, mtlsSetting MtlsConfig) error {
	// Читаем CA сертификат из файла, чтобы проверить клиентские сертификаты
	caCert, err := os.ReadFile(mtlsSetting.Cert)
	if err != nil {
		log.Printf("%s", err) // Логируем ошибку, если не удается прочитать сертификат
	}

	// Создаем пул сертификатов для проверки клиента
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Создаем TLS конфигурацию, включающую проверку клиентских сертификатов
	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,                     // Пул клиентских сертификатов для проверки
		MinVersion: tls.VersionTLS12,               // Минимальная версия TLS
		ClientAuth: tls.RequireAndVerifyClientCert, // Требуем верификацию клиентского сертификата
	}

	// Создаем HTTP сервер с TLS конфигурацией
	server := &http.Server{
		Addr:      ":9100",   // Адрес и порт для прослушивания
		TLSConfig: tlsConfig, // Устанавливаем конфигурацию TLS
	}

	// Запускаем сервер с использованием сертификата и ключа для TLS
	serverErr := server.ListenAndServeTLS(mtlsSetting.Cert, mtlsSetting.Key)
	if serverErr != nil {
		return serverErr // Если ошибка при запуске сервера, возвращаем ее
	}
	return nil
}

// RunServerWithousTls - запускает обычный HTTP сервер без поддержки mTLS
// Сервер будет слушать на порту 9100 и обрабатывать запросы, используя указанный handler.
func RunServerWithousTls(handler http.Handler) error {
	// Создаем HTTP сервер без TLS (по умолчанию)
	server := &http.Server{
		Addr:    ":9100", // Адрес и порт для прослушивания
		Handler: handler, // Обработчик запросов
	}

	// Запускаем сервер
	serverErr := server.ListenAndServe()
	if serverErr != nil {
		return serverErr // Если ошибка при запуске сервера, возвращаем ее
	}
	return nil
}
