1. Реализация сервиса osago-aggregator

1.1. Нуждается ли osago-aggregator в собственном хранилище данных

Да, требуется собственное хранилище.

Причины:
1. Асинхронный характер процесса:
* создание заявки,
* ожидание предложений до 60 секунд,
* частичное поступление результатов от разных страховых компаний.
2. Масштабирование сервиса в несколько экземпляров:
* состояние не может храниться в памяти pod’а.
3. Необходимость:
* дедупликации запросов,
* контроля таймаутов,
* повторных опросов страховых компаний,
* восстановления процесса при рестарте pod’а.

Минимальный набор данных:
* osago_request_id (внутренний)
* core_app_request_id
* данные автомобиля (нормализованные)
* список страховых компаний
* состояние по каждой СК:
* status: CREATED | IN_PROGRESS | COMPLETED | FAILED | TIMEOUT
* external_request_id
* timestamp последнего опроса
* результат (если получен)
* дедлайн (now + 60s)

Технологически:
* PostgreSQL (отдельная схема или отдельная БД)
* либо NoSQL (например, Redis + persistence), но PostgreSQL проще для консистентности и аналитики


2. API сервиса osago-aggregator

2.1. API для core-app

Синхронное REST API (request–response) + асинхронные события

1. Создание заявки

POST /osago/requests

Request:
* данные автомобиля
* идентификатор пользователя
* correlationId

Response (синхронно, < 100–200 мс):

{
  "osagoRequestId": "uuid",
  "insuranceCompanies": [
    { "companyId": "A", "status": "IN_PROGRESS" },
    { "companyId": "B", "status": "IN_PROGRESS" }
  ]
}

GET /osago/requests/{id}

Возвращает:
* все полученные предложения
* текущие статусы остальных
Используется как fallback, но не как основной механизм доставки.


3. Интеграция core-app/osago-aggregator

3.1. Основное средство интеграции
Комбинация:
 * REST - для создания заявки
 * Event-driven (message broker KAFKA) - для доставки предложений


3.2. Поток взаимодействия
1. core-app -> osago-aggregator (REST)
* создаёт заявку
2. osago-aggregator:
* создаёт записи в БД
* отправляет запросы во все страховые компании
3. По мере получения ответа от СК:
* публикует событие
4. core-app подписан на события и обновляет состояние заявки

4. Интеграция osago-aggregator / страховые компании

4.1. Характер интеграции
* Внешние REST API
* Высокая латентность
* Непредсказуемое поведение партнёров

4.2. Реализация
* Для каждой страховой компании:
* собственный client / adapter
* Вызовы:
1. POST /osago/applications
2. GET /osago/applications/{id}/offer
* Опрос (polling):
* экспоненциальный backoff
* max deadline = 60 секунд

5. API core-app для веб-приложения
WebSocket / Server-Sent Events (SSE)

REST не подходит, так как:
 * потребует агрессивного polling,
 * увеличит нагрузку при множестве одновременных пользователей.

API
1. Создание заявки
POST /api/osago/requests
2. Подписка на обновления 
GET /api/osago/requests/{id}/stream
SSE или WebSocket channel

core-app:
* принимает события от osago-aggregator,
* пушит их в браузер конкретного пользователя.

6. Применение паттернов отказоустойчивости

Web App -> core-app
* Timeout - защита UI от зависших запросов

core-app -> osago-aggregator
* Rate Limiting - защита от: злоупотреблений, багов фронта, партнёрских B2B вызовов
* Circuit Breaker - при деградации osago-aggregator
* Timeout - короткий (100–300 мс)

core-app -> client-info
* Rate Limiting - защита от: злоупотреблений, багов фронта
* Circuit Breaker - при деградации client-info
* Timeout - короткий (100–300 мс)

osago-aggregator -> страховые компании
* Circuit Breaker (per partner) - один проблемный партнёр не влияет на остальных
* Retry -только для идемпотентных операций
* Timeout - индивидуальный, но ≤ 60 секунд total

7. Учёт деплоя в нескольких экземплярах

Корректное функционирование системы возможно только при следующих условиях.
1. Нет состояния в памяти
* всё состояние в БД / брокере
2. Sticky sessions не требуются
* WebSocket:
* либо через shared session store,
* либо через message broker
3. Idempotency
* при повторных вызовах POST /osago/requests
4. Горизонтальное масштабирование
* osago-aggregator масштабируется независимо от core-app
* нагрузка пользователей → fan-out X N страховых

8. Итоговая логика архитектурного решения
* osago-aggregator - асинхронный orchestration-сервис
* WebSocket/SSE - обязательны для UX
* Event-driven интеграция - ключ к масштабированию
* Агрессивная изоляция партнёров - обязательна для SLA

