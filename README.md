
```
cyberhunt
├─ cmd
│  ├─ helper.go
│  ├─ main.go
│  ├─ middleware.go
│  └─ routes.go
├─ docker-composedb.yaml
├─ Dockerfile
├─ go.mod
├─ go.sum
├─ internal
│  ├─ database
│  │  └─ database.go
│  ├─ handlers
│  │  ├─ admin.go
│  │  ├─ auth.go
│  │  ├─ game.go
│  │  ├─ handler.go
│  │  ├─ leaderboard.go
│  │  ├─ leaderboard_sse.go
│  │  └─ seed.go
│  ├─ models
│  │  └─ models.go
│  ├─ services
│  │  ├─ admin_service.go
│  │  ├─ clue_service.go
│  │  ├─ errors.go
│  │  ├─ game_service.go
│  │  └─ group_service.go
│  └─ utils
│     └─ utils.go
├─ static
│  └─ js
│     └─ html5-qrcode.min.js
└─ templates
   ├─ admin.html
   ├─ adminLogin.html
   ├─ game.html
   ├─ leaderboard.html
   ├─ login.html
   └─ seed.html

```