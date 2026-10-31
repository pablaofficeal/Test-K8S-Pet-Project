# Kubernetes: конфигурация и секреты

[English](k8s-secrets.md) | **Русский**

Фронтенд **не** входит в кластер (он на Vercel). В кластере работают `gateway`, `auth` и `backend`. **PostgreSQL работает вне кластера** (docker compose на хосте, `192.168.0.15:56799`).

## Структура

```
k8s/
├── kustomization.yaml      # точка входа: kubectl apply -k k8s/
├── namespace.yaml          # low-code-bot-platform
├── configmap.yaml          # несекретные настройки (TTL, алгоритм, Argon2, ...)
├── serviceaccounts.yaml    # по одному на сервис, токен API не монтируется
├── networkpolicies.yaml    # default deny + gateway -> auth/backend
├── database/               # Service + EndpointSlice -> внешний PostgreSQL на хосте
├── auth/ backend/ gateway/ verify/
├── rabbitmq/               # Service + EndpointSlice -> внешний RabbitMQ на хосте
├── migrations/             # Job: `migrations up` (AutoMigrate + seed), запускается как Sync-hook Argo CD
└── secrets/
    ├── secrets.example.yaml  # шаблон с плейсхолдерами (можно коммитить)
    ├── create-secrets.sh     # собирает реальные секреты из env-файла
    └── sealed/               # опционально: SealedSecrets (можно коммитить)
```

## Что является секретом и кто его получает

| Секрет                 | Ключи                                               | Потребители            |
|------------------------|-----------------------------------------------------|------------------------|
| `database-url`         | `DATABASE_URL`                                      | auth, backend          |
| `migration-db-url`     | `MIGRATION_DB_URL`                                  | Job миграций           |
| `rabbitmq-auth-url`    | `RABBITMQ_AUTH_URL`                               | **только auth** (публикует) |
| `rabbitmq-verify-url`  | `RABBITMQ_VERIFY_URL`                             | **только verify** (читает)  |
| `verify-pepper`        | `VERIFY_CODE_PEPPER`                              | **только verify**           |
| `jwt-private-key`      | `JWT_SECRET_KEY`                                    | **только auth**        |
| `jwt-public-key`       | `JWT_PUBLIC_KEY`                                    | auth, backend, gateway |

Всё остальное (`JWT_ALGORITHM`, TTL, параметры Argon2 и т.д.) не секрет и лежит в ConfigMap `app-config`.
Приватный ключ вынесен в отдельный секрет намеренно: его монтирует только `auth`.

## Деплой

1. **Перегенерируй JWT-ключи** — см. [jwt-keys.ru.md](jwt-keys.ru.md). Реальные значения положи в локальный некоммитимый env-файл (`.env.prod`).
2. Создай секреты:

   ```bash
   k8s/secrets/create-secrets.sh --env-file .env.prod
   ```

   Для тестового/учебного кластера можно взять закоммиченный моковый `.env` как есть (`--env-file .env`). Скрипт не проверяет, моковые ли ключи, — на чём-то настоящем так делать нельзя.
3. Впиши свои имена образов в `k8s/kustomization.yaml` (`DOCKERHUB_USERNAME`) и примени:

   ```bash
   kubectl apply -k k8s/
   ```

## Прод: не держи секреты в обычном `kubectl apply`

Обычные Secret в Kubernetes — это только base64, их видит любой с правом `get secret`. Рекомендуется:

- **Sealed Secrets** (работает со скриптом): `k8s/secrets/create-secrets.sh --env-file .env.prod --seal`
  кладёт зашифрованные манифесты в `k8s/secrets/sealed/`. Их можно коммитить и применять через `kubectl apply -f k8s/secrets/sealed/`.
  Нужны контроллер Sealed Secrets и CLI `kubeseal`.
- **External Secrets Operator** + Vault / облачный менеджер секретов, если он у тебя уже есть.
- Включить **encryption at rest** для Secret в кластере и ограничить RBAC на `secrets`.

## Ротация JWT-ключей

1. Сгенерируй новую пару ([jwt-keys.ru.md](jwt-keys.ru.md)) и обнови env-файл.
2. Запусти `create-secrets.sh` заново.
3. Перезапусти сервисы: `kubectl -n low-code-bot-platform rollout restart deploy/auth deploy/backend deploy/gateway`.
4. Все выданные токены станут невалидными, пользователям придётся войти заново.

## Подтверждение почты

См. [verify-service.ru.md](verify-service.ru.md). RabbitMQ работает на хосте, как и PostgreSQL (`k8s/rabbitmq/service.yaml` связывает `low-code-bot-platform-rabbitmq:5672` с `192.168.0.15:56772`); `verify` настраивается через `k8s/verify/configmap.yaml` и пока работает с **тестовым провайдером** и фиксированным кодом.

## Argo CD

Argo CD по умолчанию исключает `Endpoints`/`EndpointSlice`, из-за чего `k8s/database/service.yaml` тихо не применяется.
Один раз примени `argocd/resource-exclusions-patch.yaml` (команда в самом файле), затем создай приложение: `kubectl apply -f argocd/application.yaml`.

## Ingress

Ingress обслуживает официальный helm-чарт `ingress-nginx`, который ставится отдельным приложением Argo CD:

```bash
kubectl apply -f argocd/ingress-nginx.yaml
```

Он слушает NodePort `30080` (HTTP) и `30444` (HTTPS) на каждой ноде. Настройки из `Infra/nginx/nginx.conf` перенесены в ConfigMap контроллера в этом же файле
(`proxy-read-timeout` 365s, `proxy-buffering` off, тело до 50M, keep-alive, gzip, JSON-лог с `request_id`, `server-tokens` off). `k8s/gateway/ingress.yaml` направляет `/` на шлюз.
WebSocket upgrade, `X-Real-IP`, `X-Forwarded-For` и `X-Request-ID` ingress-nginx обрабатывает по умолчанию. Зоны `limit_req` и `proxy_cache_path` в `nginx.conf` были объявлены, но ни в одном `location` не использовались, поэтому не переносились (у шлюза свой rate limiting).

Применяй приложение Ingress **после** того, как основное приложение синхронизирует Service шлюза как `ClusterIP`, потому что оба хотят порт `30080`.

> ⚠️ `kubernetes/ingress-nginx` заархивирован в марте 2026 и больше не получает исправлений. Для тестового кластера нормально; для чего-то настоящего планируй переход на другой контроллер.

## Миграции

`auth` и `backend` ждут появления таблиц в БД; таблицы создаёт Job `migrations` (`migrations up`).
Он запускается при каждой синхронизации (идемпотентно). Остальные команды (`status`, `down`, `seed`) запускай вручную через `kubectl run` с образом миграций.

## Заметки и допущения

- В `routes.yaml` шлюза хосты записаны как `Low-Code-Bot-Platform-*`; Service названы в нижнем регистре
  (`low-code-bot-platform-auth-service`, `...-users-backend`) — DNS не чувствителен к регистру, так что шлюз их найдёт.
- Для сервиса `bot`, на который ссылается шлюз, манифестов пока нет (исходников в репозитории нет).
- Service шлюза — `ClusterIP`; снаружи он доступен через ingress-nginx (см. «Ingress»), например `http://192.168.0.31:30080/main/api/v1/`.
- `networkpolicies.yaml` применяется: в кластере Calico (VXLAN, pod CIDR `10.244.0.0/16`). Входящий трафик по умолчанию закрыт; шлюз принимает отовсюду (нужно для NodePort), `auth`/`backend` — только от шлюза.
- Реплик по 1 и небольшие лимиты памяти, потому что на воркерах ~4 ГБ RAM.
- `DATABASE_URL` сохраняет compose-хост `low-code-bot-platform-postgresql:5432`; внутри кластера он резолвится через `k8s/database/service.yaml` в БД на хосте. Если поменяются IP хоста или `POSTGRES_PORT`, правь этот файл. Контейнер БД должен быть запущен и публиковать порт в LAN.
- CI (`.github/workflows/docker-publish.yml`) собирает и пушит `backend` (`test-k8s-api`), `auth` и `gateway` в Docker Hub. Нужны секреты репозитория `DOCKERHUB_USERNAME` и `DOCKERHUB_TOKEN`.
