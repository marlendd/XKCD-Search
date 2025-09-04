# Пример публикации и подписки

Запуск с 4х консолей:

```bash
    docker run -p 4222:4222 nats
```

```bash
    go run publisher/main.go 
```

```bash
    go run subscriber1/main.go 
```

```bash
    go run subscriber2/main.go 
```

```bash
    go run subscriber3/main.go 
```
