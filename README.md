<div align="center">

# XKCD Comics Search System

Микросервисная система для поиска по комиксам [XKCD](https://xkcd.com). Поддерживает полнотекстовый поиск, быстрый поиск по индексу, а также поиск по изображению с использованием нейросети YOLO. Все сервисы написаны на Go, общаются через gRPC, имеют высокое тестовое покрытие (>80%) и запускаются одной командой `make up`.

</div>

---

##  Демонстрация работы

<video src="https://raw.githubusercontent.com/Kseniiiiia/GoLang-Yadro-course/main/animation.mp4" controls width="800"></video>
*Краткое видео, показывающее поиск комиксов с помощью загруженного изображения*

---

##  Архитектура

Система состоит из шести микросервисов, взаимодействующих через gRPC и REST. Все сервисы находятся в папке search-services/ и контейнеризированы через Docker Compose

| Микросервис          | Назначение                                               | Порт (внутри сети) | Протокол  |
|----------------------|----------------------------------------------------------|--------------------|-----------|
| **API Gateway**      | Единая точка входа REST + JWT-аутентификация             | 28080              | HTTP/REST |
| **Frontend**         | Веб-интерфейс (HTML/templates + WebSocket)               | 28084              | HTTP/WS   |
| **Search Service**   | Полнотекстовый и индексный поиск                         | 28083              | gRPC      |
| **Update Service**   | Загрузка комиксов с https://xkcd.com/ в БД               | 28082              | gRPC      |
| **Words Normalizer** | Стемминг и фильтрация стоп-слов                          | 28081              | gRPC      |
| **Yolo Service**     | Детекция объектов на изображении (прокси к внешнему API) | 28085              | gRPC      |

---

##  Микросервисы

### 1. Words Normalizer
**Папка:** ` search-services/words/`  
**Технологии:** Snowball stemmer, стоп-слова (английские)

**Функции:**
- Принимает фразу, разбивает на слова, удаляет знаки препинания, стемминг, отбрасывает стоп-слова
- Возвращает список уникальных нормализованных слов
- Ограничение длины входной фразы – 20480 байт (при превышении возвращает `codes.ResourceExhausted`)

**gRPC API (proto/words/words.proto):**
```protobuf
service Words {
  rpc Ping(google.protobuf.Empty) returns (google.protobuf.Empty);
  rpc Norm(WordsRequest) returns (WordsReply);
}
```

**Реализация:**
- Чистая функция `Norm(phrase string) []string` без внешних зависимостей
- Использует `snowball.Stem` и `english.IsStopWord`
- Удаляет дубликаты через `map`

---

### 2. Update Service
**Папка:** ` search-services/update/`  
**Задача:** Загружает недостающие комиксы с xkcd.com, нормализует их текст и сохраняет в PostgreSQL

**Основные компоненты:**
- **core.Service** – ядро с методами:
  - `Update()` – параллельная загрузка (конкурентность задаётся в конфиге). Пропускает уже существующие ID
  - `Stats()` – возвращает статистику БД и общее количество комиксов на XKCD
  - `Status()` – текущее состояние обновления (idle/running)
  - `Drop()` – очистка таблицы.
- **Адаптеры:**
  - `db.DB` – PostgreSQL с миграциями (встроенные SQL через `embed`). Таблица: `comics (id INT PRIMARY KEY, url TEXT, words TEXT[])`.
  - `xkcd.Client` – HTTP-клиент к xkcd.com. Отслеживает `missingIDs` (404).
  - `words.Client` – gRPC-клиент к Words Normalizer.
  - `grpc.Server` – реализует методы из `proto/update.proto`: `Update`, `Status`, `Stats`, `Drop`, `Ping`.
- **Миграции:** автоматически применяются при старте (`db.Migrate()`).

**gRPC API (proto/update/update.proto):**
```protobuf
service Update {
  rpc Ping(google.protobuf.Empty) returns (google.protobuf.Empty);
  rpc Status(google.protobuf.Empty) returns (StatusReply);
  rpc Update(google.protobuf.Empty) returns (google.protobuf.Empty);
  rpc Stats(google.protobuf.Empty) returns (StatsReply);
  rpc Drop(google.protobuf.Empty) returns (google.protobuf.Empty);
}
```

**Конфигурация (config.yaml):**
```yaml
log_level: DEBUG
update_address: localhost:28082
words_address: localhost:28081
db_address: localhost:5432
xkcd:
  url: https://xkcd.com
  concurrency: 10
  timeout: 10s
  check_period: 1h
```

---

### 3. Search Service
**Папка:** ` search-services/search/`  
**Задача:** Предоставляет два режима поиска: полнотекстовый по БД и быстрый по обратному индексу. Индекс периодически перестраивается

**Основные компоненты:**
- **core.Service** – ядро:
  - `Search()` – нормализует фразу через Words, выполняет сложный SQL-запрос к PostgreSQL (ранжирование по уникальным и общим совпадениям)
  - `IndexSearch()` – использует обратный индекс в памяти (`map[word][]comicID`), сортирует по релевантности
  - `BuildIndex()` – перестраивает индекс из всех комиксов в БД
  - `Stats()` – статистика БД
- **Адаптеры:**
  - `db.DB` – PostgreSQL (такая же таблица, как в Update Service)
  - `words.Client` – gRPC-клиент к Words Normalizer
  - `grpc.Server` – реализует методы из `proto/search.proto`: `Search`, `IndexSearch`, `Ping`
  - `initiator.Initiator` – фоновый процесс, перестраивающий индекс с интервалом `index_ttl`

**gRPC API (proto/search.proto):**
```protobuf
service Search {
  rpc Search(SearchRequest) returns (SearchResponse);
  rpc IndexSearch(IndexSearchRequest) returns (SearchResponse);
  rpc Ping(google.protobuf.Empty) returns (google.protobuf.Empty);
}
```

**Конфигурация (config.yaml):**
```yaml
log_level: DEBUG
words_address: localhost:28081
db_address: localhost:5432
search_server:
    address: localhost:28083
    timeout: 10s
    index_ttl: 20s
```

**Индексация:**
- Индекс хранится в памяти под `sync.RWMutex`
- Перестраивается при старте и затем каждые `index_ttl`
- Формат: `map[string][]int` (слово → список ID комиксов)

---

### 4. API Gateway
**Папка:** ` search-services/api/`  
**Задача:** Единая точка входа для HTTP-клиентов, обеспечивает аутентификацию (JWT), rate limiting, ограничение параллельных запросов и проксирует вызовы к gRPC-сервисам

**Основные компоненты:**
- **HTTP-обработчики (rest):** `/api/login`, `/api/search`, `/api/isearch`, `/api/db/update`, `/api/db/stats`, `/api/db/status`, `/api/db` (DELETE), `/api/detect`, `/api/ping`, `/api/words`
- **Middleware:**
  - `Auth` – проверка JWT-токена (заголовок `Authorization: Token <jwt>`)
  - `Concurrency` – ограничение одновременных запросов (семафор)
  - `Rate` – ограничение запросов в секунду
- **gRPC-клиенты:** к Words, Update, Search, Yolo
- **Сервис аутентификации (aaa):** проверяет логин/пароль администратора из переменных окружения `ADMIN_USER`/`ADMIN_PASSWORD`, выдаёт JWT

**Конфигурация (config.yaml):**
```yaml
log_level: DEBUG
search_concurrency: 1
search_rate: 1
token_ttl: 1m
words_address: localhost:28081
update_address: localhost:28082
search_address: localhost:28083
yolo_address: localhost:28085
api_server:
  address: localhost:28080
  timeout: 5s
```

---

### 5. Frontend
**Папка:** ` search-services/comic-frontend/`  
**Задача:** Веб-интерфейс для пользователей. Реализован на HTML/templates, общается с API Gateway через HTTP и WebSocket

**Страницы:**
- Главная (`/`) – форма поиска с переключателем быстрого/обычного режима
- Поиск по изображению (`/image-search`) – загрузка картинки, отправка на `/detect`
- Результаты поиска (`/results`) – отображение найденных комиксов
- Админ-панель (`/admin`) – защищена JWT, отображает статистику и статус обновления, позволяет запустить обновление или сбросить БД
- Логин (`/admin/login`) – форма входа для администратора

**WebSocket:** на `/admin/update/update-progress` передаёт текущую статистику и статус обновления в реальном времени

**Конфигурация (config.yaml):**
```yaml
log_level: DEBUG
server:
  address: localhost:28084
  timeout: 2m
api:
  url: http://localhost:28080
```

---

### 6. Yolo Service
**Папка:** ` search-services/yolo/`  
**Задача:** Прокси-сервис для внешнего YOLO API. Принимает gRPC-запрос с изображением, преобразует его в формат, ожидаемый внешним HTTP-сервисом, и возвращает результаты детекции

**Особенности:**
- Принимает `DetectRequest` с бинарными данными изображения
- Кодирует изображение в base64 и оборачивает в JSON вида `{"image": {"py/b64": "..."}}`
- Отправляет POST-запрос на внешний API (адрес задаётся в конфиге как `yolo_api_address`)
- Декодирует ответ, ожидая структуру с полем `yolo_results`, содержащим `bbox`, `det_score`, `label_num`, `label_string`
- Преобразует в protobuf-ответ `DetectResponse`

**gRPC API (proto/yolo.proto):**
```protobuf
service YoloService {
  rpc Detect (DetectRequest) returns (DetectResponse);
}
message DetectRequest { bytes image_data = 1; }
message DetectResponse { repeated Detection results = 1; }
message Detection {
  repeated float bboxes = 1;
  float confidence = 2;
  string label = 3;
  int32 label_num = 4;
}
```

**Конфигурация (config.yaml):**
```yaml
yolo_address: localhost:28085
yolo_api_address: localhost:10004   # адрес внешнего YOLO HTTP API
```

---

##  База данных (PostgreSQL)

Таблица `comics` создаётся миграцией:

```sql
CREATE TABLE comics (
    id    INTEGER PRIMARY KEY,
    url   TEXT NOT NULL,
    words TEXT[]
);
```

- `id` – номер комикса на xkcd.com
- `url` – прямая ссылка на изображение
- `words` – массив нормализованных слов (из заголовка, описания и alt-текста)

---

## Индексация и поиск

- **Обратный индекс:** Search Service периодически (каждые `index_ttl`) перестраивает в памяти индекс: `слово -> список ID комиксов`. Это позволяет выполнять поиск по индексу (`IndexSearch`) в несколько раз быстрее, чем полнотекстовый запрос к БД
- **Ранжирование:** при поиске по индексу комиксы сортируются сначала по количеству уникальных совпадающих слов, затем по общему числу совпадений. При полнотекстовом поиске аналогичная логика реализована в SQL

---

##  Запуск и развертывание

### Требования
- Docker 20.10+ и Docker Compose 2.0+
- 4+ GB RAM (рекомендуется 8 GB)

### Docker Compose

Файл [`compose.yaml`](compose.yaml) описывает всю инфраструктуру проекта. Он включает:

- **6 Go-микросервисов** (`frontend`, `yolo`, `api`, `words`, `update`, `search`) – каждый собирается из своего Dockerfile (лежат в папке `search-services/`)
- **PostgreSQL** – база данных
- **pgAdmin** – веб-интерфейс для управления БД
- **YOLO API** – внешний сервис детекции [`tae898/yolov5`](https://hub.docker.com/r/tae898/yolov5)

**Ключевые особенности:**
- Сервисы связаны через внутреннюю сеть Docker
- Используются `depends_on` с `condition: service_healthy` для PostgreSQL, чтобы база была готова до старта `update` и `search`
- Проброшены конфигурационные файлы через volumes для лёгкого изменения настроек без пересборки
- Переменные окружения задают адреса зависимых сервисов

### Makefile

Проект включает [`Makefile`](Makefile) для автоматизации частых задач:

| Команда          | Описание                                                              |
|------------------|-----------------------------------------------------------------------|
| `make up`        | Собрать и запустить все сервисы                                       |
| `make down`      | Остановить все сервисы                                                |
| `make clean`     | Остановить сервисы и удалить тома (база данных будет стёрта)          |
| `make run-tests` | Запустить интеграционные тесты в отдельном контейнере                 |
| `make test`      | Полный цикл: clean -> up -> ждать -> run-tests -> clean               |
| `make lint`      | Запустить линтер для Go-кода (через `search-services/Makefile`)       |
| `make proto`     | Сгенерировать gRPC-код из protobuf (через `search-services/Makefile`) |
| `make unit`      | Запустить unit-тесты и скопировать отчёт о покрытии `cover.html`      |

### Быстрый старт

```bash
git clone https://github.com/Kseniiiiia/GoLang-Yadro-course.git
cd GoLang-Yadro-course

# Запуск всех сервисов
make up
```

После запуска:
- Фронтенд доступен по адресу `http://localhost:28084`
- API Gateway – `http://localhost:28080`

### Первоначальная настройка
1. Перейдите в админ-панель (`/admin`)
2. Войдите с учётными данными по умолчанию: `admin` / `password` (задаются через переменные окружения `ADMIN_USER`/`ADMIN_PASSWORD`)
3. Нажмите **Update Database** – загрузка всех комиксов с xkcd.com займёт около 30–40 секунд (при `concurrency: 10`)
4. После завершения можно выполнять поиск

### Переменные окружения
Основные настройки можно переопределить через переменные окружения в `compose.yaml` (см. секции `environment` у каждого сервиса).

---

## Тестирование

### Unit-тесты
Каждый микросервис покрыт unit-тестами с использованием **gomock** и **sqlxmock**. Покрытие >80% (подтверждено `cover.html` в корне репозитория).

### Интеграционные тесты
В папке `tests/` находятся интеграционные тесты, которые проверяют работу всей системы в сборе. Они используют реальные запущенные сервисы (адрес по умолчанию `http://localhost:28080`). Тесты включают:
- Проверку доступности сервисов (`/api/ping`).
- Логин и аутентификацию
- Очистку и наполнение БД
- Параллельные запросы на обновление
- Поиск по фразам (как обычный, так и индексный)
- Проверку rate limiting и ограничения конкурентности
- Истечение срока действия JWT

Запуск интеграционных тестов (требуются запущенные сервисы):
```bash
cd tests
go test -v
```

Или через:
```bash
make test
```

Некоторые тесты помечены как длительные (ждут 30 секунд и более) и могут быть пропущены с флагом `-short`.

---

## API Endpoints (через API Gateway)

| Метод    | Endpoint                            | Описание                                                     | Аутентификация |
|----------|-------------------------------------|--------------------------------------------------------------|----------------|
| `POST`   | `/api/login`                        | Получение JWT (JSON `{"name": "admin", "password": "..."}`)  | -              |
| `GET`    | `/api/ping`                         | Проверка доступности сервисов (возвращает JSON со статусами) | -              |
| `GET`    | `/api/words?phrase=...`             | Нормализация фразы (возвращает список слов)                  | -              |
| `GET`    | `/api/search?phrase=...&limit=...`  | Полнотекстовый поиск                                         | -              |
| `GET`    | `/api/isearch?phrase=...&limit=...` | Поиск по индексу (быстрый)                                   | -              |
| `POST`   | `/api/db/update`                    | Запуск обновления базы комиксов                              | (admin)        |
| `GET`    | `/api/db/stats`                     | Статистика базы (количество слов, комиксов)                  | -              |
| `GET`    | `/api/db/status`                    | Статус обновления (`idle`/`running`)                         | -              |
| `DELETE` | `/api/db`                           | Очистка базы (drop)                                          | (admin)        |
| `POST`   | `/api/detect`                       | Поиск по изображению (multipart/form-data с полем `image`)   | -              |

---

## Структура репозитория
```
.
├── search-services/          
│   ├── api/                  # API Gateway
│   ├── comic-frontend/       # Веб-интерфейс                
│   ├── proto/                # gRPC прото-файлы
│   ├── search/               # Search Service
│   ├── update/               # Update Service
│   ├── words/                # Words Normalizer
│   ├── yolo/                 # Yolo Service
│   ├── cover.html            # Отчёт о покрытии тестами
│   ├── Dockerfile.api        # Dockerfile для API Gateway
│   ├── Dockerfile.frontend   # Dockerfile для Frontend
│   ├── Dockerfile.search     # Dockerfile для Search
│   ├── Dockerfile.update     # Dockerfile для Update
│   ├── Dockerfile.words      # Dockerfile для Words
│   └── Dockerfile.yolo       # Dockerfile для Yolo
├── tests/                    # Интеграционные тесты
├── docker-compose.yaml       # Оркестрация всех сервисов
├── Makefile                  # Автоматизация команд
├── animation.mp4             # Видео-демонстрация
└── README.md                 # Документация
```