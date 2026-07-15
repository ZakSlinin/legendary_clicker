Run project:

```
docker-compose up --build
```

For frontend: 

```
POST /register
Отдаём: {"user_id": 123}
Возвращает: {"balance": 0} 
```

```
POST /tap
Отдаём: {"user_id": 123, "tap_counts": 7}
Возвращает: {"balance": 42}
```

```
GET /balance?user_id=123
Ничего не отдаём (id в query)
Возвращает: {"balance": 42}
```