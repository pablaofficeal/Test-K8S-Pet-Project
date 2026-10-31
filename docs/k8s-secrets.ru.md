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
├── auth/ backend/ gateway/
└── secrets/
    ├── secrets.example.yaml  # шаблон с плейсхолдерами (можно коммитить)
    ├── create-secrets.sh     # собирает реальные секреты из env-файла
    └── sealed/               # опционально: SealedSecrets (можно коммитить)
```

## Что является секретом и кто его получает

| Секрет                 | Ключи                                               | Потребители            |
|------------------------|-----------------------------------------------------|------------------------|
| `database-url`         | `DATABASE_URL`                                      | auth, backend          |
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

## Заметки и допущения

- В `routes.yaml` шлюза хосты записаны как `Low-Code-Bot-Platform-*`; Service названы в нижнем регистре
  (`low-code-bot-platform-auth-service`, `...-users-backend`) — DNS не чувствителен к регистру, так что шлюз их найдёт.
- Для сервиса `bot`, на который ссылается шлюз, манифестов пока нет (исходников в репозитории нет).
- Шлюз опубликован как `NodePort` (`30080`) на каждой ноде, например `http://192.168.0.31:30080/main/api/v1/`. Ingress не описан.
- `networkpolicies.yaml` применяется: в кластере Calico (VXLAN, pod CIDR `10.244.0.0/16`). Входящий трафик по умолчанию закрыт; шлюз принимает отовсюду (нужно для NodePort), `auth`/`backend` — только от шлюза.
- Реплик по 1 и небольшие лимиты памяти, потому что на воркерах ~4 ГБ RAM.
- `DATABASE_URL` сохраняет compose-хост `low-code-bot-platform-postgresql:5432`; внутри кластера он резолвится через `k8s/database/service.yaml` в БД на хосте. Если поменяются IP хоста или `POSTGRES_PORT`, правь этот файл. Контейнер БД должен быть запущен и публиковать порт в LAN.
- CI (`.github/workflows/docker-publish.yml`) собирает и пушит `backend` (`test-k8s-api`), `auth` и `gateway` в Docker Hub. Нужны секреты репозитория `DOCKERHUB_USERNAME` и `DOCKERHUB_TOKEN`.
