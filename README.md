# Orbital
Система управления отложенными сообщениями с tiered storage.

## Требования

- Работа на уровне 10-50ms
- Большая пропускная способность (1_000_000+ RPS)
- Горизонтальное масштабирование
- Отказоустойчивость
- Шардирование

---

## Архитектура

### Общая схема

```
                              ┌─────────────────────────────────────────────┐
                              │               Coordinator                   │
                              │   (конфигурация, ноды, правила, tiers)      │
                              └─────────────────────────────────────────────┘
                                        ▲            ▲            ▲
                                        │            │            │
┌──────────┐   ┌──────────┐   ┌─────────┴────────────┴────────────┴─────────┐
│ Producer │─ ▶│ Gateway  │──▶│    NATS JetStream (orbital.storage.>)       │
└──────────┘   │ (вход)   │   └─────────┬────────────┬────────────┬─────────┘
               └──────────┘             │            │            │
                                        ▼            ▼            ▼
                              ┌─────────────────────────────────────────────┐
                              │                  Storage                    │
                              │   ┌─────────┐  ┌─────────┐  ┌─────────┐     │
                              │   │  Redis  │  │Postgres │  │   S3    │     │
                              │   │  (hot)  │  │ (warm)  │  │ (cold)  │     │
                              │   └────┬────┘  └────┬────┘  └────┬────┘     │
                              └────────┼────────────┼────────────┼──────────┘
                                       │            │            │
                                       ▼            ▼            ▼
                              ┌─────────────────────────────────────────────┐
                              │    NATS JetStream (orbital.ready.>)         │
                              └─────────────────────────────────────────────┘
                                                    │
                                                    ▼
                              ┌─────────────────────────────────────────────┐
                              │                 Gateway                     │
                              │     (выход → проверка правил → пушеры)      │
                              └─────────────────────────────────────────────┘
                                                    │
                                                    ▼
                              ┌─────────────────────────────────────────────┐
                              │    NATS JetStream (orbital.push.>)          │
                              └─────────┬────────────┬────────────┬─────────┘
                                        │            │            │
                                        ▼            ▼            ▼
                              ┌──────────┐    ┌──────────┐    ┌──────────┐
                              │ Pusher 1 │    │ Pusher 2 │    │ Pusher N │
                              │  (HTTP)  │    │ (Kafka)  │    │  (gRPC)  │
                              └──────────┘    └──────────┘    └──────────┘
```

### NATS как универсальный буфер

NATS JetStream используется как буфер **между всеми компонентами** системы.
Все взаимодействия между сервисами идут через NATS — прямых HTTP-вызовов
между Gateway, Storage и Pusher нет.

### NATS Subjects

| Subject | Назначение | Producer | Consumer |
|---------|------------|----------|----------|
| `orbital.storage.<storage_id>` | Сообщения для конкретного storage | Gateway | Storage |
| `orbital.promote.<storage_id>` | Продвижение сообщений в storage | Storage (верхний tier) | Storage (нижний tier) |
| `orbital.ready` | Готовые к отправке сообщения | All Storages | Gateway |
| `orbital.push.<pusher_id>` | Сообщения для конкретного пушера | Gateway | Pusher |

Subject для storage формируется из ID, который задаётся при регистрации (см. [Именование Storage](#именование-storage)).

### Поток сообщения

1. **Producer** отправляет сообщение в **Gateway**
2. **Gateway** определяет storage по `ScheduledAt` (сравнивает с `MinDelay`/`MaxDelay` зарегистрированных storages) и публикует в NATS:
   - `< 1 мин` → `orbital.storage.hot-l1`
   - `1 мин - 1 час` → `orbital.storage.warm-l1`
   - `> 1 час` → `orbital.storage.cold-l1`
3. **Storage** получает из NATS и сохраняет сообщение
4. При приближении времени Storage публикует в `orbital.promote.<storage_id>` нижнего tier'а
5. Когда `ScheduledAt` наступает, Storage публикует в `orbital.ready`
6. **Gateway** получает из `orbital.ready`, применяет **RoutingRules**
7. **Gateway** публикует в `orbital.push.<pusher_id>` для каждого совпавшего пушера
8. **Pusher** получает из NATS и отправляет во внешнюю систему

---

## Компоненты

### Message

Основная единица данных в системе.

```go
type Message struct {
    ID          string            // Уникальный идентификатор (автогенерация)
    RoutingKey  string            // Ключ маршрутизации
    Payload     []byte            // Полезная нагрузка
    Metadata    map[string]string // Метаданные
    CreatedAt   time.Time         // Время создания
    ScheduledAt time.Time         // Время доставки
}
```

**Опции создания:**
| Опция | Описание |
|-------|----------|
| `WithID(string)` | Установить ID (для восстановления) |
| `WithRoutingKey(string)` | Ключ маршрутизации |
| `WithPayload([]byte)` | Полезная нагрузка |
| `WithMetadata(map[string]string)` | Метаданные целиком |
| `WithMetadataValue(key, value)` | Одна пара ключ-значение |
| `WithScheduledAt(time.Time)` | Точное время доставки |
| `WithDelay(time.Duration)` | Задержка от текущего момента |

---

### Gateway

Точка входа и выхода сообщений.

```go
type Gateway interface {
    Consume(message *Message) error
}
```

**Обязанности:**
- Принимает сообщения от producers
- Определяет tier по `ScheduledAt` (запрос к Coordinator)
- Направляет в соответствующий Storage
- При истечении времени — применяет RoutingRules и отправляет в Pushers

---

### MessageStorage

Интерфейс хранилища сообщений. Реализации: Redis, PostgreSQL, S3.

```go
type MessageStorage interface {
    Store(ctx, msg) error                           // Сохранить
    FetchExpiring(ctx, threshold, limit) ([]*StoredMessage, error)  // Истекающие
    FetchReady(ctx, limit) ([]*StoredMessage, error) // Готовые к отправке
    Acknowledge(ctx, msgID) error                   // Подтвердить обработку
    Reject(ctx, msgID, requeue) error               // Отклонить
    Get(ctx, msgID) (*StoredMessage, error)         // Получить по ID
    Delete(ctx, msgID) error                        // Удалить
    Count(ctx) (int64, error)                       // Количество
    Close() error                                   // Закрыть
}
```

**StoredMessage** — обёртка с метаданными:
```go
type StoredMessage struct {
    *Message
    Status        MessageStatus  // Pending, InFlight, Delivered, Failed
    Attempts      int            // Количество попыток
    LastAttemptAt time.Time      // Время последней попытки
}
```

---

### Coordinator

Центральный компонент управления кластером. Предоставляет service discovery и конфигурацию для всех компонентов системы.

**Архитектура:**

```
┌─────────────────────────────────────────────────────────┐
│                      etcd cluster                        │
│   (распределённое хранилище, Raft консенсус)            │
└─────────────────────────────────────────────────────────┘
            ▲                           ▲
            │                           │
   ┌────────┴────────┐         ┌────────┴────────┐
   │  Coordinator 1  │         │  Coordinator 2  │
   │    (активный)   │         │   (реплика)     │
   └────────┬────────┘         └────────┬────────┘
            │                           │
            └───────────┬───────────────┘
                        │
        ┌───────────────┼───────────────┐
        ▼               ▼               ▼
   ┌─────────┐    ┌──────────┐    ┌─────────┐
   │ Gateway │    │ Storage  │    │ Pusher  │
   └─────────┘    └──────────┘    └─────────┘
```

**Роль координатора:**

| Функция | Описание |
|---------|----------|
| Service Discovery | Регистрация и обнаружение Gateways, Storages, Pushers |
| Конфигурация | Хранение RoutingRules, адреса NATS, общих настроек |
| Storage Selection | Определение хранилища по задержке сообщения |
| Health Monitoring | Отслеживание heartbeat компонентов (фоновая задача) |
| Cleanup | Удаление мёртвых нод (фоновая задача) |

**Хранилище (etcd):**

| Ключ | Данные |
|------|--------|
| `/orbital/nodes/{id}` | Ноды координатора |
| `/orbital/gateways/{id}` | Gateway инстансы |
| `/orbital/storages/{id}` | Storage инстансы с диапазонами задержек |
| `/orbital/pushers/{id}` | Pusher инстансы |
| `/orbital/routing-rules/{id}` | Правила маршрутизации |
| `/orbital/config` | Общая конфигурация |
| `/orbital/nats-address` | Адрес NATS сервера |

**Интерфейс CoordinatorStorage:**

```go
type CoordinatorStorage interface {
    // Coordinator Nodes
    CreateNode(ctx, node) error
    GetNode(ctx, nodeID) (*Node, error)
    ListNodes(ctx) ([]*Node, error)
    UpdateNodeHeartbeat(ctx, nodeID) error
    DeleteNode(ctx, nodeID) error

    // Gateways
    RegisterGateway(ctx, gateway) error
    ListGateways(ctx) ([]*GatewayInfo, error)
    // ...

    // Storages
    RegisterStorage(ctx, storage) error
    FindStorageForDelay(ctx, delay) (*StorageInfo, error)
    // ...

    // Pushers, RoutingRules, Config, NATS...
}
```

**Реализации:**
- `internal/coordinator/storage/etcd` — production (etcd backend)

---

### Coordinator Node

Представляет инстанс координатора в кластере. Используется для:
- Отслеживания живых координаторов
- Leader election (один координатор выполняет фоновые задачи)
- Health checks между координаторами

```go
type Node struct {
    id            uuid.UUID
    address       string
    status        NodeStatus  // Connecting, Active, Removed
    registeredAt  time.Time
    lastHeartbeat time.Time
}
```

**Статусы ноды:**

| Статус | Описание |
|--------|----------|
| `Connecting` | Нода запускается, ещё не готова |
| `Active` | Нода работает, heartbeat актуален |
| `Removed` | Нода удалена из кластера |

**Методы:**
- `IsAlive(timeout)` — проверка что heartbeat свежий
- `IsActive()` — проверка статуса Active

---

### Зарегистрированные компоненты

При старте каждый компонент регистрируется в координаторе:

**GatewayInfo:**
```go
type GatewayInfo struct {
    ID            string
    Address       string
    Status        NodeStatus
    RegisteredAt  time.Time
    LastHeartbeat time.Time
}
```

**StorageInfo:**
```go
type StorageInfo struct {
    ID       string          // Человекочитаемый ID (см. Именование Storage)
    Address  string
    MinDelay time.Duration   // Минимальная задержка
    MaxDelay time.Duration   // Максимальная задержка (0 = без ограничения)
    Status   NodeStatus
    // ...
}
```

Storage регистрирует диапазон задержек, которые он обслуживает.
ID используется как часть NATS subject: `orbital.storage.{ID}`.

Пример регистрации:

| ID | Хранилище | MinDelay | MaxDelay | NATS subject |
|----|-----------|----------|----------|--------------|
| `hot-l1` | Redis | 0 | 1m | `orbital.storage.hot-l1` |
| `warm-l1` | PostgreSQL | 1m | 1h | `orbital.storage.warm-l1` |
| `cold-l1` | S3 | 1h | 0 (unlimited) | `orbital.storage.cold-l1` |

**PusherInfo:**
```go
type PusherInfo struct {
    ID      string
    Type    string  // "http", "kafka", "grpc", "nats"
    Address string
    Status  NodeStatus
    // ...
}
```

---

### Pusher

Отправляет сообщения во внешние системы.

```go
type Pusher interface {
    Push(msg *Message) error
}
```

**Реализации (планируемые):**
- HTTP Webhook
- Kafka
- gRPC
- NATS
- Custom

---

### RoutingRule (планируется)

Правила маршрутизации сообщений к пушерам.

| Тип | Описание | Пример |
|-----|----------|--------|
| `MatchExact` | Точное совпадение | `orders` |
| `MatchPrefix` | Начинается с | `orders.` |
| `MatchSuffix` | Оканчивается на | `.eu` |
| `MatchRegex` | Регулярное выражение | `^orders\..*` |

---

## NATS JetStream

### Почему NATS

| Преимущество | Для Orbital |
|--------------|-------------|
| Низкая latency (~100μs) | Критично для 1M+ RPS |
| JetStream persistence | At-least-once delivery |
| Consumer groups | Масштабирование consumers |
| Back pressure | Защита от перегрузки |
| Subject-based routing | Гибкая маршрутизация |

### Streams

| Stream | Subjects | Retention | Назначение |
|--------|----------|-----------|------------|
| `ORBITAL_STORAGE` | `orbital.storage.>` | WorkQueue | Входящие сообщения для storage |
| `ORBITAL_PROMOTE` | `orbital.promote.>` | WorkQueue | Продвижение между tiers |
| `ORBITAL_READY` | `orbital.ready` | WorkQueue | Готовые к отправке |
| `ORBITAL_PUSH` | `orbital.push.>` | WorkQueue | Отправка в пушеры |

### Consumer Groups

Каждый сервис при старте подписывается на свой NATS subject по ID:

```
Stream: ORBITAL_STORAGE
├── Consumer: hot-l1     (filter: orbital.storage.hot-l1)
├── Consumer: warm-l1    (filter: orbital.storage.warm-l1)
└── Consumer: cold-l1    (filter: orbital.storage.cold-l1)

Stream: ORBITAL_PUSH
├── Consumer: pusher-http-1
├── Consumer: pusher-http-2    (масштабирование)
└── Consumer: pusher-kafka
```

### Гарантии доставки

| Этап | Гарантия | Механизм |
|------|----------|----------|
| Gateway → Storage | At-least-once | NATS Ack |
| Storage → Gateway | At-least-once | NATS Ack |
| Gateway → Pusher | At-least-once | NATS Ack |
| Pusher → External | Depends on pusher | Retry + DLQ |

### Пример публикации

```go
// Gateway публикует в storage по ID
js.Publish("orbital.storage.hot-l1", msgData)

// Storage публикует готовое сообщение
js.Publish("orbital.ready", msgData)

// Gateway публикует в пушер
js.Publish("orbital.push.http-webhook-1", msgData)
```

---

## Tiered Storage

Количество уровней хранения **не ограничено**. Каждый storage регистрируется
в координаторе с уникальным ID и диапазоном задержек. Gateway автоматически
направляет сообщения в подходящий storage.

### Пример с 3 уровнями

| Tier | ID | Хранилище | Задержка | Характеристики |
|------|----|-----------|----------|----------------|
| **Hot** | `hot-l1` | Redis | < 1 мин | Высокая скорость, ограниченная память |
| **Warm** | `warm-l1` | PostgreSQL | 1 мин - 1 час | Надёжность, SQL-запросы |
| **Cold** | `cold-l1` | S3 | > 1 час | Дёшево, большой объём |

### Пример с 5 уровнями

| Tier | ID | Хранилище | Задержка |
|------|----|-----------|----------|
| **Hot** | `hot-l1` | Redis | < 30с |
| **Hot** | `hot-l2` | Redis Cluster | 30с - 5 мин |
| **Warm** | `warm-l1` | PostgreSQL | 5 мин - 1 час |
| **Warm** | `warm-l2` | PostgreSQL (archive) | 1 час - 24 часа |
| **Cold** | `cold-l1` | S3 | > 24 часа |

### Именование Storage

Формат ID: **`{tier}-l{level}`**

| Часть | Описание | Примеры |
|-------|----------|---------|
| `tier` | Температурный класс хранилища | `hot`, `warm`, `cold` |
| `l{level}` | Уровень внутри tier'а (1 = ближе к доставке) | `l1`, `l2`, `l3` |

- `hot-l1` — самый быстрый, сообщения вот-вот уйдут
- `warm-l1` — первый уровень тёплого хранения
- `cold-l2` — второй уровень холодного хранения (более долгосрочный)

Ограничения: ID должен содержать только строчные латинские буквы, цифры и дефис (`^[a-z0-9][a-z0-9-]*$`), чтобы быть валидным токеном NATS subject.

**Продвижение сообщений:**
```
cold-l1 ──[осталось < 1 час]──▶ warm-l1 ──[осталось < 1 мин]──▶ hot-l1
```

---

## Структура проекта

```
orbital/
├── cmd/                              # Точки входа
│   ├── gateway/main.go               # Gateway сервис
│   ├── coordinator/main.go           # Coordinator сервис
│   ├── storages/
│   │   ├── storage-redis/main.go     # Redis storage
│   │   ├── storage-postgres/main.go  # PostgreSQL storage
│   │   └── storage-s3/main.go        # S3 storage
│   └── all-in-one/main.go            # Всё в одном (dev)
│
├── internal/                         # Внутренние реализации
│   └── coordinator/
│       └── storage/
│           └── etcd/
│               ├── storage.go        # etcd реализация CoordinatorStorage
│               └── dto.go            # DTO для сериализации
│
├── pkg/                              # Публичные пакеты
│   ├── nats/
│   │   └── client.go                # NATS/JetStream клиент с логированием
│   ├── sdk/
│   │   ├── coordinator/client.go    # SDK для координатора
│   │   └── storage/client.go        # SDK для отправки в storage через NATS
│   └── entities/
│       ├── message.go                # Message + опции
│       ├── gateway.go                # Gateway interface
│       ├── pusher.go                 # Pusher interface
│       ├── event_type.go             # EventType enum
│       ├── storage/
│       │   └── storage.go            # MessageStorage interface
│       ├── routing_rule/
│       │   └── routing_rule.go       # RoutingRule struct
│       └── coordinator/
│           ├── coordinator.go        # Coordinator interface
│           ├── storage.go            # CoordinatorStorage interface
│           ├── node.go               # Node struct
│           └── config.go             # CoordinatorConfig
│
├── deploy/
│   └── docker/
│       ├── Dockerfile.gateway
│       ├── Dockerfile.coordinator
│       ├── Dockerfile.storage-redis
│       ├── Dockerfile.all-in-one
│       └── docker-compose.yml
│
├── go.mod
└── README.md
```

---

## Запуск

### Development (all-in-one)

```bash
go run ./cmd/all-in-one
```

### Docker Compose

```bash
cd deploy/docker
docker-compose up -d
```

### Отдельные сервисы

```bash
# Coordinator
docker build -f deploy/docker/Dockerfile.coordinator -t orbital-coordinator .
docker run -p 8081:8080 orbital-coordinator

# Gateway
docker build -f deploy/docker/Dockerfile.gateway -t orbital-gateway .
docker run -p 8080:8080 -e COORDINATOR_ADDR=coordinator:8080 orbital-gateway
```

---

## Конфигурация

Через переменные окружения:

| Переменная | Описание | Пример |
|------------|----------|--------|
| `NATS_URL` | URL NATS сервера | `nats://nats:4222` |
| `NATS_CREDS` | Путь к credentials файлу | `/etc/nats/creds` |
| `COORDINATOR_ADDR` | Адрес coordinator | `coordinator:8080` |
| `ETCD_ENDPOINTS` | Адреса etcd | `etcd:2379` |
| `REDIS_ADDR` | Адрес Redis | `redis:6379` |
| `POSTGRES_DSN` | DSN PostgreSQL | `postgres://...` |
| `S3_ENDPOINT` | S3 endpoint | `s3.amazonaws.com` |

---

## TODO

- [ ] NATS JetStream setup (streams, consumers)
- [ ] Реализация Gateway (publish/subscribe)
- [ ] Реализация MessageStorage (Redis, PostgreSQL, S3)
- [x] Реализация CoordinatorStorage с etcd backend
- [ ] HTTP API для Coordinator
- [ ] Leader election для координаторов
- [ ] RoutingRules matcher
- [ ] HTTP/gRPC API для producers
- [ ] Метрики (Prometheus)
- [ ] Трейсинг (OpenTelemetry)
- [ ] Dead Letter Queue (NATS stream)
- [ ] Retry политики
- [ ] Web UI для мониторинга
