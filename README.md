# LightChat

A minimalist full-stack real-time messenger with user authentication, direct and group chats, WebSocket messaging, message search, and dark mode.

---

## Stack

| Layer     | Technology                                 |
|-----------|--------------------------------------------|
| frontend  | vue 3, typescript, vite, tailwind css      |
| state     | pinia, vue router                          |
| backend   | go, chi, sqlx, golang-jwt, bcrypt          |
| realtime  | websockets                                 |
| database  | postgresql                                 |
| infra     | docker compose                             |

---

## Features

User Management
- Secure auth: registration, login, logout
- Refresh token rotation: automatic access token renewal via HTTP-only cookies
- Token reuse detection: revokes all sessions if a rotated refresh token is replayed
- Profile editing: display name, username, email, status

Chats
- Direct messages: find users by exact username, open a private chat in one click
- Group chats: create with a custom name, add and remove members
- Resizable sidebar with compact icon mode, swipe gestures on mobile
- Chat deletion with confirmation

Messaging
- Send and receive text messages instantly via WebSockets
- Automatic reconnect with exponential backoff on connection loss
- Read receipts: messages marked as read when the recipient opens the chat
- Edit and delete your own messages, changes broadcast to all participants in real time
- Infinite scroll: load older messages on scroll-up with position preservation
- Server-side message search: full-text ILIKE search via dedicated API endpoint

UI & Theme
- Dark mode: light/dark theme switching
- Responsive layout: works on desktop and mobile

---

## Getting started
```bash
git clone https://github.com/Andrii-K-17/light-chat.git
cd light-chat
```
```bash
cp .env.example .env
```
```bash
docker-compose up -d --build
```

open `http://localhost:5173`

---

## Project structure
```
light-chat
├── backend/                  # Go backend
│   ├── cmd/                  # Entry points
│   │   └── server/           # Main server executable
│   └── internal/             # Core application logic
│       ├── config/           # Configuration
│       ├── db/               # Database connection
│       ├── handlers/         # HTTP request handlers
│       ├── middleware/       # JWT auth middleware
│       ├── models/           # Data models & structs
│       ├── repository/       # Persistence layer
│       ├── response/         # Standardized API responses
│       ├── router/           # Route definitions
│       ├── services/         # Business logic & JWT helpers
│       └── ws/               # WebSocket hub & connection handler
│
├── frontend/                 # Vue 3 + TypeScript frontend
│   └── src/
│       ├── api/              # API calls to backend
│       ├── components/       # Reusable UI components
│       │   ├── chat/         # Chat-specific components
│       │   └── ui/           # Generic UI components
│       ├── router/           # Vue Router configuration
│       ├── stores/           # Pinia state management
│       ├── types/            # TypeScript interfaces
│       └── views/            # Page-level components
│
└── init.sql                  # Database schema
```

---

## API
```
POST   /api/auth/register                    create account
POST   /api/auth/login                       authenticate
POST   /api/auth/logout                      end session
POST   /api/auth/refresh                     rotate token pair

GET    /api/auth/me                          current user
PATCH  /api/auth/me                          update profile

GET    /api/users/search?username=           find user by exact username

GET    /api/chats                            list chats with last message & unread count
POST   /api/chats                            create direct or group chat
DELETE /api/chats/:id                        delete chat

GET    /api/chats/:id/members                list group members
POST   /api/chats/:id/members                add member by username (creator only)
DELETE /api/chats/:id/members/:memberId      remove member (creator only)

GET    /api/chats/:id/messages               paginated message history
GET    /api/chats/:id/messages/search?q=     full-text message search
PATCH  /api/chats/:id/messages/:messageId    edit message (owner only)
DELETE /api/chats/:id/messages/:messageId    delete message (owner only)

WS     /ws?chat_id=                          real-time WebSocket connection
```

---

## WebSocket events

| Direction       | Type              | Description                        |
|-----------------|-------------------|------------------------------------|
| client → server | `send_message`    | send a new message                 |
| client → server | `read_receipt`    | mark messages as read              |
| server → client | `new_message`     | new message broadcast to chat      |
| server → client | `message_updated` | edited message broadcast to chat   |
| server → client | `message_deleted` | deleted message broadcast to chat  |
| server → client | `read_receipt`    | read receipt broadcast to chat     |

---

## Environment
```bash
ENV=development
PORT=8080
ALLOWED_ORIGIN=http://localhost:5173

DB_HOST=db
DB_PORT=5432
DB_NAME=lightchat_db
DB_USER=appuser
DB_PASSWORD=strongpassword
DB_SSLMODE=disable

JWT_SECRET=change-this-to-a-long-secret
JWT_EXPIRY_MINUTES=15
REFRESH_EXPIRY_DAYS=30
```

---

## License

This project is licensed under the [MIT License](LICENSE).
