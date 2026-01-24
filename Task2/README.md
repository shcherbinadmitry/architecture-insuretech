# Task 2: Динамическое масштабирование контейнеров

## Описание

Настройка Horizontal Pod Autoscaler (HPA) для автоматического масштабирования приложения на основе утилизации памяти.

scaletestapp-6659db8585-tdn6q## Сборка образа для Apple Silicon (ARM64)

Оригинальный образ `ghcr.io/yandex-practicum/scaletestapp` не поддерживает ARM64. Для Apple Silicon необходимо собрать образ локально:

```bash
# Переключиться на Docker внутри Minikube
eval $(minikube docker-env)

# Собрать образ
docker build -t scaletestapp:local Task2/app/

# Вернуться к локальному Docker (опционально)
eval $(minikube docker-env -u)
```

В deployment.yaml используется:
- `image: scaletestapp:v4`
- `imagePullPolicy: Never` - чтобы Kubernetes не пытался скачать образ из registry

## Структура файлов

```
Task2/
├── deployment/
│   ├── deployment.yaml    # Манифест Deployment
│   ├── service.yaml       # Манифест Service
│   └── hpa.yaml          # Манифест HorizontalPodAutoscaler
├── locustfile.py         # Сценарий нагрузочного тестирования
├── screenshots/          # Скриншоты результатов
└── README.md
```

## Инструкция по развёртыванию

### 1. Запуск Minikube

```bash
minikube start
```

### 2. Активация metrics-server

```bash
minikube addons enable metrics-server
```

### 3. Применение манифестов

```bash
# Применить все манифесты
kubectl apply -f Task2/deployment/

# Или по отдельности:
kubectl apply -f Task2/deployment/deployment.yaml
kubectl apply -f Task2/deployment/service.yaml
kubectl apply -f Task2/deployment/hpa.yaml
```

### 4. Проверка статуса

```bash
# Проверить deployment
kubectl get deployments

# Проверить pods
kubectl get pods

# Проверить service
kubectl get services

# Проверить HPA
kubectl get hpa

# Детальная информация о HPA
kubectl describe hpa scaletestapp-hpa
```

### 5. Получение URL сервиса

```bash
minikube service scaletestapp-service --url
```

## Нагрузочное тестирование

### 1. Установка Locust

```bash
pip install locust
```

### 2. Запуск Locust

```bash
cd Task2
locust
```

### 3. Настройка теста

1. Откройте http://localhost:8089 в браузере
2. Укажите URL приложения (результат команды `minikube service scaletestapp-service --url`)
3. Установите количество пользователей (например, 100)
4. Установите Spawn rate (например, 10)
5. Нажмите "Start swarming"

### 4. Мониторинг масштабирования

```bash
# Наблюдение за HPA в реальном времени
kubectl get hpa -w

# Наблюдение за подами
kubectl get pods -w

# Открыть дашборд Kubernetes
minikube dashboard
```

## Параметры конфигурации

### Deployment
- **Replicas**: 1 (начальное количество)
- **Memory limit**: 30Mi
- **Memory request**: 20Mi
- **Image**: ghcr.io/yandex-practicum/scaletestapp:latest
- **Port**: 8080

### HPA
- **Min replicas**: 1
- **Max replicas**: 10
- **Target memory utilization**: 80%

## Ожидаемое поведение

1. При низкой нагрузке работает 1 реплика
2. При увеличении нагрузки и превышении 80% утилизации памяти HPA создаёт дополнительные реплики
3. Максимальное количество реплик - 10
4. При снижении нагрузки количество реплик постепенно уменьшается

## Очистка ресурсов

```bash
kubectl delete -f Task2/deployment/
minikube stop
```
