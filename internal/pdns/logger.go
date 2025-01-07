package pdns

import (
	"log/slog"   // Импортируем стандартный логгер из пакета log/slog для логирования
	"log/syslog" // Импортируем пакет для работы с системным журналом syslog
	"os"         // Импортируем пакет для работы с операционной системой, в частности с файлами
)

func initLogger(logPath, logLevel string, LogToSyslog, LogToFile bool) {
	// Устанавливаем уровень логирования (DEBUG, INFO, WARN, ERROR)
	// В зависимости от переданного logLevel, определяется уровень детализации логирования
	var level slog.Level
	var src bool
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug // Для уровня DEBUG включаем более подробное логирование
		src = true              // Включаем информацию о файле и строке (источник)
	case "INFO":
		level = slog.LevelInfo // Для уровня INFO логируются основные сообщения
		src = false            // Не добавляем информацию о файле и строке
	case "WARN":
		level = slog.LevelWarn // Для уровня WARN логируются предупреждения
		src = false            // Не добавляем информацию о файле и строке
	case "ERROR":
		level = slog.LevelError // Для уровня ERROR логируются только ошибки
		src = false             // Не добавляем информацию о файле и строке
	}

	// Переменная для хранения обработчика логов, который будет использоваться для записи сообщений
	var handler slog.Handler
	var err error

	// Проверяем, что параметры логирования корректно настроены:
	// Если оба флага (LogToFile и LogToSyslog) установлены в true или оба в false,
	// выводим предупреждающее сообщение о неправильной настройке.
	if LogToFile && LogToSyslog {
		// Если оба флага установлены в true, выводим предупреждение.
		slog.Warn("The logging parameters are incorrectly configured (logToSyslog and logToFile should not be set to both false or both true; by default, syslog is used).")
	} else if !LogToFile && !LogToSyslog {
		// Если оба флага установлены в false, выводим предупреждение.
		slog.Warn("The logging parameters are incorrectly configured (logToSyslog and logToFile should not be set to both false or both true; by default, syslog is used).")
	}

	// Используем switch для выбора логирования в зависимости от значений LogToSyslog и LogToFile
	// Важно, чтобы хотя бы один из флагов был установлен в true.
	switch {
	case LogToSyslog:
		// Логируем только в syslog
		// Создаем обработчик для syslog с уровнем LOG_INFO и меткой "dns-monitor"
		syslogHandler, err := syslog.New(syslog.LOG_INFO, "dns-monitor")
		if err != nil {
			// Если произошла ошибка при создании обработчика, логируем ошибку
			slog.Error(err.Error())
		}
		// Настроим обработчик для записи в syslog в формате JSON
		handler = slog.NewJSONHandler(syslogHandler, &slog.HandlerOptions{
			AddSource: src,   // Добавляем информацию о файле и строке в лог
			Level:     level, // Устанавливаем уровень логирования
		})

	case LogToFile:
		// Логируем только в файл
		// Открываем файл для записи, создаем его, если он не существует, и добавляем новые записи в конец
		var logFile *os.File
		logFile, err = os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			// Если произошла ошибка при открытии файла, логируем ошибку
			slog.Error(err.Error())
		}
		// Настроим обработчик для записи в файл в формате JSON
		handler = slog.NewJSONHandler(logFile, &slog.HandlerOptions{
			AddSource: src,   // Добавляем информацию о файле и строке в лог
			Level:     level, // Устанавливаем уровень логирования
		})

	default:
		// Если оба флага false, используем syslog по умолчанию
		// Создаем обработчик для syslog с уровнем LOG_INFO и меткой "dns-monitor"
		syslogHandler, err := syslog.New(syslog.LOG_INFO, "dns-monitor")
		if err != nil {
			// Если произошла ошибка при создании обработчика, логируем ошибку
			slog.Error(err.Error())
		}
		// Настроим обработчик для записи в syslog в формате JSON
		handler = slog.NewJSONHandler(syslogHandler, &slog.HandlerOptions{
			AddSource: src,   // Добавляем информацию о файле и строке в лог
			Level:     level, // Устанавливаем уровень логирования
		})
	}

	// Создаем новый логгер с выбранным обработчиком
	var logger = slog.New(handler)

	// Добавляем теги к логгеру. Теги позволяют фильтровать и классифицировать логи.
	// Пример:
	// - "component" может использоваться для указания компонента приложения
	// - "environment" может использоваться для указания окружения (например, "production" или "development")
	logger = logger.With(
		slog.String("component", "dns-monitor"),  // Добавляем тег для компонента приложения
		slog.String("environment", "production"), // Добавляем тег для окружения (например, production, development)
	)

	// Устанавливаем созданный логгер как дефолтный для всей программы
	slog.SetDefault(logger)
}
