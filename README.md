# BannerService

## Инструкция по запуску
### Поднятие контейнеров go, redis и субд Postgres:
```shell
docker-compose up -d
```

## Проблемы и их решения
**1. Однозначное определние баннера по тегу и фиче**

Для решения данной проблемы было установлено условие, что не может быть баннеров с одинаковыми фичамии тегами, которые пересекаются в одном теге.
Но баннеры с разными фичами могут существовать независимо от наличия одинаковых тегов.

**3. Как получать неактуальные баннеры в обход БД**

Было принято решение использовать кэш, чтобы при получении очередного баннера из бд, он кэшировался. Это позволяет ускорить работу системы, но ведёт к потере актуальности данных.
В качестве кэша был выбран Redis.

**4. Как реализовать получение актуальной информации для 10% пользователей**

Каждый запрос на получение баннера получает свой номер на уровне middleware. Далее, если номер запроса кратен 10, пользователь получает информацию напрямуб из БД.

**5. Автоматическое применение миграций**

Для системы конроля миграций была выбрана golang библиотека goose. Во время сборки контейнера go, устанавливаются необходимые миграции и осщуствляется применение миграция.
В случае, если необходимо применить новые миграции, следует пересобрать контейнер.

**6. Система авторизации и аутентификации по токенам**

На мой взгляд, в реальных условиях выгоднее иметь отдельный сервис авторизации пользователей, который, в свою очередь генерирует токены. Используя такой токен, пользователь может проходить аутентификацию в других связанных сервисах. Поэтому, в представленном сервисе баннеров отсутсвует регистрация пользователей и их авторизация. В данном случае присутсвует лишь тестова генерация токенов дял проверки функциональности.
