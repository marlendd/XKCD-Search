# Управление нагрузкой и доступом

## Цель

Ограничить нагрузку на сервис с помощью rate и concurrency limiters. Добавить контроль доступа
к критическим операциям.

В данном задании необходимо познакомиться с шаблоном Адаптер (middleware) в REST API.

Для endpoint-а /api/search, который ходит напрямую в базу данных, нужно добавить
concurrency limiter, заданный переменной окружения SEARCH_CONCURRENCY. При достижении лимита
на работу с данным endpoint-ом система должна возвращать HTTP статус 503 (service unavailable).

Для endpoint-а /api/isearch нужно добавить rate limiter, который регулируется переменной
окружения SEARCH_RATE. Данный limiter не возвращает 503, а задерживает соединения для регулировки
их скорости в пределах заданного переменной окружения RPS (requests per second).

Два критических endpoint-а должны быть защищены от доступа непривилегированными пользователями.
Доступ регулируется через 'POST /api/login' endpoint и middleware для проверки выданных логином
JWT токенов. "POST /api/login" должен принимать JSON в теле запроса вида

```json
{
  "name": "admin",
  "password": "password"
}
```

и отдавать токен ввиде строчки. Вам ненужно реализовывать отдельный микросервис авторизации -
достаточно принимать из переменных среды имя и пароль "суперпользователя", переменные
ADMIN_USER и ADMIN_PASSWORD. "/api/login" проверяет пользователя и пароль, и если они не совпадают,
отдаем HTTP Unauthorized. Если все ОК - выдаем токен на время TOKEN_TTL (переменная среды,
2 минуты по умолчанию) с subject установленным в "superuser". При запросах на обновление базы
или ее удаление, необходимо реализовать middleware, котрое будет выданный токен проверять -
HTTP Header вида "Authorization: Token выданный_токен_здесь". Если не удалось проверить токен -
HTTP Unauthorized.

Предлагается сервис проверки и выдачи токенов реализовать в адаптере - это позволит потом плавно
перейти к разработке отдельного сервиса AAA (Authentication, Authorization, Accounting).

Сервисы должны собираться и запускаться через модифицированный compose файл,
а также проходить интеграционные тесты - запуск специального тест контейнера.

## Критерии приемки

1. Микросервисы компилируются в docker image-ы, запускаются через compose файл и проходят тесты.
2. Можно использовать код из предыдущего задания.
3. Сервис api конфигурируeтся через cleanenv пакет и должeн уметь запускаться как с config.yaml
файлом через флаг -config, так и через переменные среды, в этом задании -
ADMIN_USER, ADMIN_PASSWORD, TOKEN_TTL, API_ADDRESS, WORDS_ADDRESS, UPDATE_ADDRESS,
SEARCH_ADDRESS, SEARCH_CONCURRENCY, SEARCH_RATE. Все они уже добавлены в compose.yaml.
4. Используется golang 1.25+, slog логгер.

## Материалы для ознакомления

Rate limiters:

- [System Design - ограничитель трафика](http://youtube.com/watch?v=w4suQQtnYmY)
- [Go rate limiter](https://github.com/uber-go/ratelimit)
- [x/time/rate](https://pkg.go.dev/golang.org/x/time/rate)

Middleware:

- [Logging middleware example](https://gowebexamples.com/basic-middleware/)
- [Разработка REST-серверов на Go. Часть 5: Middleware](https://habr.com/ru/companies/ruvds/articles/566198/)

JWT:

- [golang-jwt/jwt](https://pkg.go.dev/github.com/golang-jwt/jwt/v5)
- [A guide to JWT authentication in Go](https://blog.logrocket.com/jwt-authentication-go/)
- [JWT-авторизация на сервере](https://ru.hexlet.io/courses/go-web-development/lessons/auth/theory_unit)
