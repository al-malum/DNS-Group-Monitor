package pdns

import (
	"crypto/tls"    // Пакет для работы с TLS (Transport Layer Security)
	"crypto/x509"   // Пакет для работы с сертификатами X.509
	"encoding/json" // Пакет для работы с JSON
	"log/slog"      // Логирование с использованием slog
	"net/http"      // Пакет для создания HTTP серверов
	"os"            // Пакет для работы с операционной системой
)

// AuthenticationCN - middleware для проверки мTLS аутентификации с использованием Common Name (CN)
func AuthenticationCN(next http.Handler, mtlsSetting MtlsConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что запрос был отправлен через mTLS (проверка наличия верифицированной цепочки сертификатов)
		if r.TLS != nil && len(r.TLS.VerifiedChains) > 0 && len(r.TLS.VerifiedChains[0]) > 0 {
			var commonName = r.TLS.VerifiedChains[0][0].Subject.CommonName // Извлекаем Common Name (CN) из сертификата клиента
			// Если CN есть в списке разрешенных CN в конфиге, то аутентификация успешна
			if ContainString(mtlsSetting.AllowedCN, commonName) {
				slog.Info("Authentication successful",
					slog.String("CN", commonName),
					slog.String("remoteAddr", r.RemoteAddr))
				next.ServeHTTP(w, r) // Передаем запрос следующему обработчику
			} else if len(mtlsSetting.AllowedCN) == 0 || !mtlsSetting.Enabled {
				// Если мTLS не включен или список разрешенных CN пуст, разрешаем запрос
				slog.Info("mTLS disabled, request without mTLS successful",
					slog.String("remoteAddr", r.RemoteAddr))
				next.ServeHTTP(w, r)
			} else {
				// Если CN не разрешен, возвращаем ошибку 403 (Forbidden)
				slog.Warn("Authentication failed - incorrect CN",
					slog.String("CN", commonName),
					slog.String("remoteAddr", r.RemoteAddr))
				w.WriteHeader(http.StatusForbidden)
				w.Header().Set("Content-Type", "application/json")
				response := make(map[string]string)
				response["message"] = "Incorrect CN of the certificate"
				jsonResponse, _ := json.Marshal(response) // Преобразуем ответ в JSON
				w.Write(jsonResponse)
			}
		} else {
			// Если mTLS не используется, разрешаем запрос
			slog.Info("mTLS disabled, request without mTLS successful",
				slog.String("remoteAddr", r.RemoteAddr))
			next.ServeHTTP(w, r)
		}
	})
}

// RunServerWithTls - запускает HTTPS сервер с поддержкой mTLS
func RunServerWithTls(handler http.Handler, mtlsSetting MtlsConfig) error {
	// Логируем начало процесса запуска сервера с mTLS
	slog.Info("Starting HTTPS server with mTLS",
		slog.String("cert", mtlsSetting.Cert),
		slog.String("key", mtlsSetting.Key))

	// Читаем CA сертификат из файла, чтобы проверить клиентские сертификаты
	caCert, err := os.ReadFile(mtlsSetting.Cert)
	if err != nil {
		slog.Error("Failed to read CA cert file",
			slog.String("cert", mtlsSetting.Cert),
			slog.String("error", err.Error()))
		return err // Логируем ошибку и возвращаем её
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
		slog.Error("Error starting HTTPS server",
			slog.String("error", serverErr.Error()))
		return serverErr // Логируем ошибку при старте и возвращаем её
	}

	// Логируем успешный запуск сервера
	slog.Info("HTTPS server started successfully", slog.String("addr", server.Addr))
	return nil
}

// RunServerWithousTls - запускает обычный HTTP сервер без поддержки mTLS
func RunServerWithousTls(handler http.Handler) error {
	// Логируем начало процесса запуска сервера без mTLS
	slog.Info("Starting HTTP server without mTLS", slog.String("addr", ":9100"))

	// Создаем HTTP сервер без TLS (по умолчанию)
	server := &http.Server{
		Addr:    ":9100", // Адрес и порт для прослушивания
		Handler: handler, // Обработчик запросов
	}

	// Запускаем сервер
	serverErr := server.ListenAndServe()
	if serverErr != nil {
		slog.Error("Error starting HTTP server",
			slog.String("error", serverErr.Error()))
		return serverErr // Логируем ошибку при старте и возвращаем её
	}

	// Логируем успешный запуск сервера
	slog.Info("HTTP server started successfully", slog.String("addr", server.Addr))
	return nil
}
