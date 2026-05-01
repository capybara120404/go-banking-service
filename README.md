# go-banking-service

REST API для банковского сервиса на языке Go, приложение реализует основные функции банкинга, включая регистрацию и аутентификацию пользователей, управление счетами и картами, переводы, оформление кредитов и расчет графиков платежей, а также аналитику.

## Особенности

- **Слоистая архитектура**: приложение разделено на основные слои: модели, репозитории, сервисы и обработчики.

- **Безопасность**:
  - Шифрование данных банковских карт через PGP и HMAC.
  - Хеширование паролей с использованием bcrypt.
  - Хеширование CVV с использованием bcrypt.
  - Аутентификация с использованием JWT.

- **Интеграции**:
  - Получение ключевой ставки ЦБ РФ через SOAP-запросы с использованием `beevik/etree`.
  - Отправка email-уведомлений о платежах через SMTP с использованием `go-mail/mail/v2`.

- **База данных**: PostgreSQL с использованием транзакций.

- **Логирование**: Использование библиотеки `logrus`.

- **Автообработчик**: Автоматическая обработка платежей по кредитам.

## Технологии

- Go 1.23+
- PostgreSQL
- gorilla/mux (маршрутизатор)
- golang-jwt/jwt/v5 (JWT аутентификация)
- sirupsen/logrus (логирование)
- go-mail/mail/v2 (SMTP)
- beevik/etree (XML/SOAP парсинг)
- golang.org/x/crypto (bcrypt)

## Требования

- Установить Go 1.23+.
- Установить PostgreSQL с расширением `pgcrypto`.

## Переменные окружения

Перед запуском приложения необходимо настроить переменные окружения и создать файлы pgp_public.key и pgp_private.key в которые положить публичный и приватный pgp ключи.

Пример файла .env:

```env
# Сервер
PORT=8080

# База данных
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=banking_db
DB_SSLMODE=disable

# Безопасность
JWT_SECRET=your_jwt_secret_key_here
HMAC_SECRET=your_hmac_secret_key_here

# SMTP Настройки
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=your_email@example.com
SMTP_PASS=your_email_password
```

Пример pgp_public.key

```text
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: OpenPGP.js v4.10.10
Comment: https://openpgpjs.org

xsBNBGn02NQBCACsBk9Y8ovcf2sl8e+Bia/VG8gvcEQCMnKqKELfYz2BmTm5
iqQvgffQ0BKYIQPyf/a0reeozwlbmX1lCRUzvRRmBI6hBziX3KFlWdxSZscU
l+0g3F3S35m+GUqe1oRymC8SSEcehcAOk1YC8rdklLVJoR0407ewuL8EAA1+
rk4KsDqYZmJ+sCFvqhUHF6S4EmQWjrntsw8jIud2HZIf7/dmjmoDcNowxdj9
gyFGOEOzXWF/gZ2qr/ZVELpeD1prKu36y8LuRa7OTf/rFoe58IPa5eRW+nBT
I9C+8YOfomCZEDrxEPxqqyaa7emg/H7puXsXWE26lJZvF0Wbdu9b4gmNABEB
AAHNFXRlc3QgPHRlc3RAdGVzdC50ZXN0PsLAjQQQAQgAIAUCafTY1AYLCQcI
AwIEFQgKAgQWAgEAAhkBAhsDAh4BACEJEOq8gTUSslW4FiEEGPwzVqcwc3+M
/3dR6ryBNRKyVbhwIwgAnDBcFTxnKw1mRRDlrJNRgdQhjW0u+VLDY8Ypg8DF
Z70le63DjpVUapIGUjeKrmsT4sILKEgK0YmejocUzvhfkFImduFGirRPYGvT
75Aa/3p6eNUxnbzNvMIMBnEcX447gc3OpfJfDA3C+xYJgfx03uhVzJRxKYcD
gGMyY+MCv7zBaReB8IKjG4Gfd92A3E0YJfhqP5bPeH7DaiW7F8uFeMMdBMN5
A9iWz7dI+L1pqIWFjv7/FUqmVV7tAR6bKczjk7KWuk+QfwCM8hUs0yeDrWMX
6zfvDaGidDdJ7tEXgXNzz2cVuNZPGITOlNN4k/VES/UjGH0NQniOJk0Nzznl
fs7ATQRp9NjUAQgA0ScTPESVilTHLL49AHiJqLsnEpC0A412VXpeU9/yhcj0
4YAp0YsaszNgLqR4dN49i8Q9RBW6zaMbNBOEoOKJne0KmYJYhH7jguKIX2do
etVgtBZXaVT2dyPgbeV3z0n+QHIoLRp3Y0VsN+H5G4Yzkp0V/97rAqCPIMAk
cFAyQ4/F6m+slCTRY+XS2cLkds+1IV7B097S+TryBaUJhULeXTV2EByP/suh
IipuwQnYFJg6OWFxqpWk/DrEz2v/D+OFBRcfKxVRhd/hStjHQXGmWZyUjWOx
WgCyD03cYgrW9tXRsRl+QjyNMdpkBAEyYgzD22FiloRNb/5UBRH0NZ+NfQAR
AQABwsB2BBgBCAAJBQJp9NjUAhsMACEJEOq8gTUSslW4FiEEGPwzVqcwc3+M
/3dR6ryBNRKyVbj4Hgf/WWYtKLb3sehMgiD8NWtCgjDBYnGmK4zw7taK9ITT
QIqOBfxxTraxN7L431P4Qiu0GjP6zrcecrek3uxzNJ1YehN/ZYoRy5LPBUns
QWje0MO4ANetvsaCD6Ca/ymrt3s0GHQP58jIaZQxCvTuWvLp5e1ztoDTHpVa
mn3/wAxWewy6tYqa46QLBcLnUgsxzM61LbjsD2kgq4XYtMFxlRaml8/21vhy
xRFrcfMKsCjZ2Fr2FKleNPktKKg0G0krpZ2Yha46R8YqV7bhYHI8kAwxvT5M
uNsvksgjdVZlxOKWzrK+kXPJ9pMsSbBQ5TFvZDmzVdtWteCw+rglwYMfI2Pi
PA==
=KjmQ
-----END PGP PUBLIC KEY BLOCK-----
```

Пример pgp_private.key

```text
-----BEGIN PGP PRIVATE KEY BLOCK-----
Version: OpenPGP.js v4.10.10
Comment: https://openpgpjs.org

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
=kh5F
-----END PGP PRIVATE KEY BLOCK-----
```

## Запуск приложения

1. Клонируйте репозиторий и перейдите в папку:

    ```bash
    git clone https://github.com/capybara120404/go-banking-service.git

    cd go-banking-service
    ```

2. Поднимите БД и примените миграции:

   ```bash
   psql -U postgres -d banking_db -f migrations/20260428145650_init_schema.sql
   ```

3. Скачайте зависимости:

   ```bash
   go mod download
   ```

4. Запустите сервер:

   ```bash
   go run cmd/api/main.go
   ```

## Структура проекта

```text
├── api_test.http        // файл с тестами для api
├── cmd
│   └── api
│       └── main.go      // Точка входа в приложение с инициализацией всех зависимостей
├── go.mod               // Файл управления зависимостями
├── go.sum               // Файл управления зависимостями
├── internal
│   ├── config           // Загрузка настроек из ENV
│   │   └── config.go
│   ├── database         // Обертка над sql.DB, для управления соединением с бд
│   │   └── storage.go
│   ├── dto              // DTO для API (Request/Response)
│   │   ├── account_response.go
│   │   ├── analytics_response.go
│   │   ├── auth_response.go
│   │   ├── card_response.go
│   │   ├── create_account_request.go
│   │   ├── create_card_request.go
│   │   ├── create_credit_request.go
│   │   ├── credit_response.go
│   │   ├── error_response.go
│   │   ├── login_request.go
│   │   ├── payment_schedule_response.go
│   │   ├── predict_balance_response.go
│   │   ├── register_request.go
│   │   └── transfer_request.go
│   ├── handler          // HTTP-обработчики
│   │   ├── account_handler.go
│   │   ├── analytics_handler.go
│   │   ├── card_handler.go
│   │   ├── credit_handler.go
│   │   ├── transaction_handler.go
│   │   └── user_handler.go
│   ├── logger           // Глобальный логгер на базе logrus
│   │   └── logger.go
│   ├── middleware       // AuthMiddleware (проверка JWT) и LoggingMiddleware
│   │   ├── auth_middleware.go
│   │   └── logging_middleware.go
│   ├── model            // Слой моделей с Go-структурами, для таблиц PostgreSQL
│   │   ├── account.go
│   │   ├── card.go
│   │   ├── credit.go
│   │   ├── payment_schedule.go
│   │   ├── transaction.go
│   │   └── user.go
│   ├── repository       // Слой репозиториев для работы с бд
│   │   ├── account_repository.go
│   │   ├── card_repository.go
│   │   ├── credit_repository.go
│   │   ├── payment_schedule_repository.go
│   │   ├── transaction_repository.go
│   │   └── user_repository.go
│   └── service          // Слой сервисов с бизнес-логикой
│       ├── account_service.go
│       ├── analytics_service.go
│       ├── card_service.go
│       ├── credit_service.go
│       ├── email_service.go
│       ├── key_rate_provider.go
│       ├── transaction_service.go
│       ├── user_service.go
│       └── utils.go
├── LICENSE
├── migrations           // Миграции для бд созданные через утилиту goose
│   └── 20260428145650_init_schema.sql
└── README.md            // Описание самого проекта
```

## Доступные эндпоинты

### Публичные

- `POST /register` - Регистрация нового пользователя
- `POST /login` - Авторизация пользователя

### Защищенные (требуют `Authorization: Bearer <token>`)

- `POST /accounts` - Создать счет
- `GET /accounts` - Получить список счетов
- `POST /accounts/deposit` - Пополнить баланс
- `POST /accounts/withdraw` - Списать средства
- `POST /cards` - Выпустить виртуальную карту
- `GET /cards` - Получить список своих карт
- `POST /transfer` - Перевод средств
- `POST /credits` - Оформление кредита
- `GET /credits/{creditId}/schedule` - График платежей по кредиту
- `GET /analytics` - Аналитика доходов и расходов пользователя
- `GET /accounts/{accountId}/predict` - Прогноз баланса на N дней

## Тестирование

API можно протестировать при помощи файла `api_test.http` в корне проекта (нужен плагин REST Client для VS Code).
