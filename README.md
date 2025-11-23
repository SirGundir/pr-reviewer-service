# PR Reviewer Assignment Service

Сервис автоматического назначения ревьюеров на Pull Request’ы внутри команды.

## Описание

Сервис позволяет:

- Управлять командами и участниками
- Создавать Pull Request’ы и автоматически назначать до 2 ревьюверов из команды автора
- Переназначать ревьюверов на других участников команды
- Получать список PR’ов, назначенных конкретному пользователю
- Помечать PR как MERGED (идемпотентно)

Сервис полностью реализован через HTTP API (OpenAPI спецификация).

---

## Стек

- Язык: Go 1.23
- База данных: PostgreSQL (через Docker)
- Миграции: встроенные (при `docker-compose up`)
- Нагрузочное тестирование: k6
- Docker / Docker Compose

---

## Запуск сервиса

1. Клонируем репозиторий:

```bash
git clone <repo_url>
cd pr-reviewer-service
```

Поднимаем сервис через Docker Compose:
```bash
docker-compose up
```


Сервис доступен на порту 8080:
```bash
curl http://localhost:8080/health
```

## API
OpenAPI спецификация находится в openapi.yaml.

## Эндпоинты:
- POST /team/add — создать команду с участниками
- GET /team/get?team_name=xxx — получить команду
- POST /team/deactivate - деактивация команды
- POST /users/setIsActive — изменить активность пользователя
- GET /users/getReview?user_id=xxx — получить PR’ы пользователя
- POST /pullRequest/create — создать PR
- POST /pullRequest/merge — пометить PR как MERGED
- POST /pullRequest/reassign — переназначить ревьювера
- GET /stats/prs - статистика по pull requests
- GET /stats/users - статистика по пользователям