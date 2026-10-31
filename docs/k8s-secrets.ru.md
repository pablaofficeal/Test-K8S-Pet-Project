# Kubernetes: конфигурация и секреты

[English](k8s-secrets.md) | **Русский**

Фронтенд **не** входит в кластер (он на Vercel). В кластере работают `gateway`, `auth`, `backend` и PostgreSQL.

## Структура

```
k8s/
├── kustomization.yaml      # точка входа: kubectl apply -k k8s/
├── namespace.yaml          # low-code-bot-platform
├── configmap.yaml          # несекретные настройки (TTL, алгоритм, Argon2, ...)
├── serviceaccounts.yaml    # по одному на сервис, токен API не монтируется
├── networkpolicies.yaml    # default deny + gateway -> auth/backend -> postgres
├── postgres/ auth/ backend/ gateway/
└── secrets/
    ├── secrets.example.yaml  # шаблон с плейсхолдерами (можно коммитить)
    ├── create-secrets.sh     # собирает реальные секреты из env-файла
    └── sealed/               # опционально: SealedSecrets (можно коммитить)
```

## Что является секретом и кто его получает

| Секрет                 | Ключи                                               | Потребители            |
|------------------------|-----------------------------------------------------|------------------------|
| `postgres-credentials` | `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | postgres               |
| `database-url`         | `DATABASE_URL`                                      | auth, backend          |
| `jwt-private-key`      | `JWT_SECRET_KEY`                                    | **только auth**        |
| `jwt-public-key`       | `JWT_PUBLIC_KEY`                                    | auth, backend, gateway |

Всё остальное (`JWT_ALGORITHM`, TTL, параметры Argon2 и т.д.) не секрет и лежит в ConfigMap `app-config`.
Приватный ключ вынесен в отдельный секрет намеренно: его монтирует только `auth`.

## Деплой

1. **Перегенерируй JWT-ключи** — см. [jwt-keys.ru.md](jwt-keys.ru.md). Реальные значения положи в локальный env-файл (например `.env`, с реальными секретами его не коммитить).
2. Создай секреты:

   ```bash
   k8s/secrets/create-secrets.sh --env-file .env
   ```

   Скрипт откажется работать, если `JWT_SECRET_KEY` совпадает с публичным моковым ключом из `.env.example`
   (`--allow-mock` отключает проверку, только для одноразовых кластеров).
3. Впиши свои имена образов в `k8s/kustomization.yaml` (`DOCKERHUB_USERNAME`) и примени:

   ```bash
   kubectl apply -k k8s/
   ```

## Прод: не держи секреты в обычном `kubectl apply`

Обычные Secret в Kubernetes — это только base64, их видит любой с правом `get secret`. Рекомендуется:

- **Sealed Secrets** (работает со скриптом): `k8s/secrets/create-secrets.sh --env-file .env --seal`
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
- `networkpolicies.yaml` применяется: в кластере Calico (VXLAN, pod CIDR `10.244.0.0/16`). Входящий трафик по умолчанию закрыт; шлюз принимает отовсюду (нужно для NodePort), `auth`/`backend` — только от шлюза, Postgres — только от `auth`/`backend`.
- Реплик по 1 и небольшие лимиты памяти, потому что на воркерах ~4 ГБ RAM. Postgres на `local-path`, поэтому его данные привязаны к одной ноде.
- `postgresql.conf` из `Infra/` не монтируется, используется стандартный конфиг PostgreSQL.
- CI сейчас собирает только `backend` (образ `test-k8s-api`); образы `auth` и `gateway` тоже нужно собирать и пушить.
