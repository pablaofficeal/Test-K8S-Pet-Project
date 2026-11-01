# Сервис подтверждения почты

[English](verify-service.md) | **Русский**

`verify/` — микросервис на Go, который подтверждает почту пользователя 6-значным кодом. Транспорт — RabbitMQ.

> ⚠️ **Сейчас активен тестовый провайдер.** Пока нет реального почтового провайдера, у всех аккаунтов один фиксированный код
> **`893658`** (`VERIFY_PROVIDER=test`). Ничего не отправляется, и любой, кто знает код, «подтвердит» любой адрес.
> Не считай подтверждённые адреса настоящими, пока тестовый провайдер не заменён, и пока не опирайся на подтверждение для контроля доступа.

## Как это работает

```
регистрация ─► auth ──(verify.events / verification.requested)──► RabbitMQ ──► verify ──► provider.Send()
                                                                                │
браузер ──► шлюз ──► verify  POST /verify/confirm {"code": "893 658"} ──────────┘ ставит users.email_verified
```

1. `auth` создаёт пользователя и публикует `{"event":"verification.requested","user_uuid":...,"trace_id":...}` в обменник `verify.events`.
   Недоступность брокера не ломает регистрацию (пользователь нажмёт «отправить ещё раз»).
2. `verify` читает очередь `verify.requests`, создаёт код (хранится только HMAC-хэш) и передаёт его провайдеру.
3. Пользователь вводит код на странице `/verify`. После успеха `users.email_verified = true`.
4. Смена почты в настройках сбрасывает подтверждение (`email_verified = false`); код, выданный на старый адрес, новый не подтвердит.

## HTTP API (через шлюз, нужен JWT)

| Метод | Путь | Результат |
|-------|------|-----------|
| `GET`  | `/main/api/v1/verify/status`  | `{verified, email, resend_in, expires_in}` |
| `POST` | `/main/api/v1/verify/resend`  | `202` новый код, `409` уже подтверждено, `429` пауза / часовой лимит |
| `POST` | `/main/api/v1/verify/confirm` | тело `{"code":"893658"}` (пробелы и дефисы игнорируются): `200`, `400` неверный код (`attempts_left`) или нет активного кода, `429` слишком много попыток |

## Лимиты (настраиваются)

| Переменная | По умолчанию | Смысл |
|------------|--------------|-------|
| `VERIFY_CODE_TTL` | `10m` | время жизни кода |
| `VERIFY_RESEND_COOLDOWN` | `60s` | минимальная пауза между кодами |
| `VERIFY_MAX_ATTEMPTS` | `5` | неверных попыток на код, потом код блокируется (даже верный отклоняется) |
| `VERIFY_MAX_SENDS_PER_HOUR` | `5` | кодов на пользователя в час |

Дополнительно шлюз ограничивает `/verify/*` по IP.

## Провайдеры

`providers.Provider` — маленький интерфейс (`Name()` и `Send(ctx, Message)`). Чтобы добавить реальный (SMTP, почтовый API и т. д.):

1. Реализуй его в `verify/providers/` и зарегистрируй в `providers/provider-new.go`.
2. Его секреты клади в Secret, **никогда не в репозиторий** (репозиторий публичный).
3. Укажи имя в `VERIFY_PROVIDER`. Коды автоматически станут случайными 6-значными, `VERIFY_TEST_CODE` будет игнорироваться.

Сервис не стартует с `VERIFY_PROVIDER=test`, если не задано `VERIFY_ALLOW_TEST_CODE=true`, и пишет предупреждение при каждом старте.

## Роли RabbitMQ

RabbitMQ работает на хосте (`docker compose up -d rabbitmq`) и описан в `Infra/rabbitmq/definitions.json`.
Vhost `cadence`, обменники и очереди объявлены заранее, поэтому приложениям не нужно право `configure`.

| Пользователь | configure | write | read | Кто использует |
|--------------|-----------|-------|------|----------------|
| `admin` | `.*` | `.*` | `.*` | люди (management UI на `127.0.0.1:${RABBITMQ_MANAGEMENT_PORT}`) |
| `auth` | – | `^verify\.events$` | – | `auth` только публикует |
| `verify` | – | – | `^verify\.requests$` | `verify` только читает |

`verify.requests` — quorum-очередь с `x-delivery-limit: 5`; «ядовитые» сообщения и те, что не обработались 5 раз,
уходят в dead-letter-очередь `verify.requests.dead` (через `verify.dead`).

Пароли в `.env` и `definitions.json` — публичные **моковые** значения. Настоящие делай через
`Infra/rabbitmq/hash-password.sh 'пароль'` (печатает хэш для `definitions.json`), а реальные URL держи в `.env.prod`.

## Запуск

- Локально: `docker compose up -d postgres rabbitmq`, миграции (`migrations up`), затем `go run .` в `verify/` с
  `DATABASE_URL`, `JWT_PUBLIC_KEY`, `RABBITMQ_VERIFY_URL` и переменными `VERIFY_*`.
- Кластер: `k8s/verify/` (Deployment, Service, ConfigMap), `k8s/rabbitmq/` (Service + EndpointSlice на брокер на хосте).
  Секреты `rabbitmq-auth-url`, `rabbitmq-verify-url` и `verify-pepper` создаёт `k8s/secrets/create-secrets.sh`.

## Ещё не сделано

- Неподтверждённым пользователям ничего не запрещено. Для этого нужен claim `email_verified` в JWT, который проверяет шлюз.
- Реальный почтовый провайдер и шаблоны писем.

## Если что-то не работает

| Симптом | Причина / решение |
|---------|-------------------|
| Под `CreateContainerConfigError`, `secret "rabbitmq-verify-url" not found` | Секретов нет в git. Запусти `k8s/secrets/create-secrets.sh --env-file .env`. |
| В логе вечно `tables not ready` | Не отработал Job `migrations`. Он идёт в sync-wave `-1`, перед Deployment, при каждой синхронизации Argo CD. |
| В логе `rabbitmq consumer stopped ... connection refused` | Порт AMQP не проброшен на хост. В `docker-compose.yml` у `rabbitmq` должен быть `"${RABBITMQ_PORT}:5672"` (одного порта управления мало). |
| Регистрация проходит, но код не приходит, в логе `auth` `RABBITMQ_AUTH_URL is empty` | `auth` запустился раньше, чем появился секрет `rabbitmq-auth-url` (он необязательный). `kubectl -n cadence-cloud rollout restart deploy/auth`. |

