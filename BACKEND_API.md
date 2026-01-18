# OM Messenger Backend API Documentation

This document lists all available API endpoints and WebSocket events for the OM Messenger backend.

## Base URL
`http://<host>:8080/api`

## Authentication

### Register
Create a new user account.
- **Endpoint**: `POST /auth/register`
- **Body**:
  ```json
  {
    "username": "johndoe",
    "email": "john@example.com",
    "password": "securepassword",
    "full_name": "John Doe"
  }
  ```
- **Response**:
  ```json
  {
    "user": { ...user object... },
    "access_token": "jwt_token",
    "refresh_token": "refresh_token"
  }
  ```

### Login
Authenticate an existing user.
- **Endpoint**: `POST /auth/login`
- **Body**:
  ```json
  {
    "email": "john@example.com",
    "password": "securepassword"
  }
  ```
- **Response**:
  ```json
  {
    "user": { ...user object... },
    "access_token": "jwt_token",
    "refresh_token": "refresh_token"
  }
  ```

### Refresh Token
Get a new access token using a refresh token.
- **Endpoint**: `POST /auth/refresh`
- **Cookie**: `om_refresh=<refresh_token>` (or handled via body if modified)
- **Response**:
  ```json
  {
    "user": { ...user object... }
  }
  ```
  *(Note: Sets new cookies/tokens)*

### Logout
Invalidate the current session.
- **Endpoint**: `POST /auth/logout`
- **Response**: `200 OK`

### CSRF Token
Get CSRF token for web clients.
- **Endpoint**: `GET /auth/csrf`
- **Response**: `{"csrf_token": "..."}`

---

## Users

### Check Username Availability
Check if a username is taken.
- **Endpoint**: `GET /users/check-username?username=johndoe`
- **Response**:
  ```json
  {
    "available": true
  }
  ```

### Get Current User
Get profile of the logged-in user.
- **Endpoint**: `GET /users/me`
- **Headers**: `Authorization: Bearer <token>`
- **Response**: `{"user": { ... }}`

### Update Profile
Update user details.
- **Endpoint**: `PUT /users/me`
- **Headers**: `Authorization: Bearer <token>`
- **Body**:
  ```json
  {
    "username": "newname",
    "full_name": "New Name"
  }
  ```
- **Response**: `{"user": { ... }}`

### Search Users
Search for users by name or username.
- **Endpoint**: `GET /users/search?q=john&limit=20`
- **Headers**: `Authorization: Bearer <token>`
- **Response**: `{"users": [ ... ]}`

### Get User by Username
Get public profile of a specific user.
- **Endpoint**: `GET /users/:identifier`
- **Notes**:
  - If `:identifier` is numeric, it is treated as a user ID (e.g. `/users/42`).
  - Otherwise it is treated as a username (e.g. `/users/johndoe`).
- **Headers**: `Authorization: Bearer <token>`
- **Response**: `{"user": { ... }}`

---

## Messages (REST)

### List Conversations
Fetch the authenticated user's conversation list (DM + groups), including the last message and unread count.

- **Endpoint**: `GET /conversations?limit=50`
- **Headers**: `Authorization: Bearer <token>`
- **Cursor pagination** (optional):
  - `cursor_created_at`: RFC3339/RFC3339Nano timestamp (the last message's `created_at` from the previous page)
  - `cursor_message_id`: uint message ID (the last message's `id` from the previous page)

Example:
`GET /conversations?limit=50&cursor_created_at=2026-01-07T12:00:00.000Z&cursor_message_id=1234`

- **Response**:
  ```json
  {
    "conversations": [
      {
        "conversation_id": "user_2",
        "peer": { "id": 2, "username": "bob", "email": "bob@example.com", "full_name": "Bob", "avatar": "", "is_online": false, "last_seen": null },
        "group": null,
        "unread_count": 3,
        "last_activity": "2026-01-07T12:34:56Z",
        "last_message": {
          "id": 1234,
          "client_id": "uuid-v4",
          "sender_id": 2,
          "sender": { "id": 2, "username": "bob", "email": "bob@example.com", "full_name": "Bob", "avatar": "", "is_online": false, "last_seen": null },
          "recipient_id": 1,
          "group_id": null,
          "content": "Hey",
          "message_type": "text",
          "status": "sent",
          "is_delivered": true,
          "is_read": false,
          "created_at": "2026-01-07T12:34:56Z",
          "created_at_unix": 1767789296
        }
      },
      {
        "conversation_id": "group_10",
        "peer": null,
        "group": { "id": 10, "name": "My Group", "icon": "", "member_count": 3 },
        "unread_count": 5,
        "last_activity": "2026-01-07T12:35:10Z",
        "last_message": {
          "id": 1240,
          "client_id": "uuid-v4",
          "sender_id": 3,
          "sender": { "id": 3, "username": "alice", "email": "alice@example.com", "full_name": "Alice", "avatar": "", "is_online": true, "last_seen": null },
          "recipient_id": null,
          "group_id": 10,
          "content": "Hello group",
          "message_type": "text",
          "status": "sent",
          "is_delivered": true,
          "is_read": false,
          "created_at": "2026-01-07T12:35:10Z",
          "created_at_unix": 1767789310
        }
      }
    ],
    "count": 2,
    "next_cursor_created_at": "2026-01-07T12:34:56.000Z",
    "next_cursor_message_id": 1234
  }
  ```

**Ordering**: Conversations are returned newest-first by `last_activity` then `message_id`. Use `next_cursor_created_at` + `next_cursor_message_id` to paginate deterministically.

### Get Messages (DM)
Fetch message history for a direct conversation.
- **Endpoint**: `GET /messages?recipient_id=123&limit=50`
- **Headers**: `Authorization: Bearer <token>`
- **Ordering**: Newest-first by `id`.
- **Response**: `{"messages": [ ... ]}`

### Send Message (REST Fallback)
Send a message via HTTP (prefer WebSocket).
- **Endpoint**: `POST /messages`
- **Headers**: `Authorization: Bearer <token>`
- **Body**:
  ```json
  {
    "recipient_id": 123,
    "content": "Hello!",
    "message_type": "text"
  }
  ```
- **Response**: `Message Object`

---

## Groups

---

## Groups

### Group Object
```json
{
  "id": 10,
  "name": "My Group",
  "description": "Cool people only",
  "icon": "",
  "creator_id": 1,
  "is_public": true,
  "handle": "mygroup",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

### Create Group
- **Endpoint**: `POST /groups`
- **Headers**: `Authorization: Bearer <token>`
- **Body**:
  ```json
  {
    "name": "My Group",
    "description": "Cool people only",
    "is_public": true,
    "handle": "mygroup"
  }
  ```
- **Response**: `Group Object`

### Get My Groups
List groups the user belongs to.
- **Endpoint**: `GET /groups`
- **Headers**: `Authorization: Bearer <token>`
- **Response**: `[ ...Group Objects... ]`

### Join Group
- **Endpoint**: `POST /groups/:id/join`
- **Headers**: `Authorization: Bearer <token>`
- **Response**: `{"message": "Joined group successfully"}`

### Search Public Groups
- **Endpoint**: `GET /groups/public/search?q=<query>&limit=20`
- **Headers**: `Authorization: Bearer <token>`
- **Notes**: Only public groups appear. Private groups are excluded.
- **Response**:
  ```json
  { "groups": [ ...Group Objects... ] }
  ```

### Get Public Group by Handle
- **Endpoint**: `GET /groups/handle/:handle`
- **Headers**: `Authorization: Bearer <token>`
- **Response**: `Group Object`

### Join Public Group by Handle
- **Endpoint**: `POST /groups/handle/:handle/join`
- **Headers**: `Authorization: Bearer <token>`
- **Response**: `Group Object`

### Create Invite Link
- **Endpoint**: `POST /groups/:id/invite-links`
- **Headers**: `Authorization: Bearer <token>`
- **Body**:
  ```json
  {
    "single_use": false,
    "expires_in_seconds": 3600
  }
  ```
- **Response**:
  ```json
  {
    "token": "abc123...",
    "join_path": "/join/abc123...",
    "join_url": "https://your-domain/join/abc123...",
    "expires_at": "2025-01-01T01:00:00Z",
    "max_uses": null
  }
  ```

### Join Group by Invite Link
- **Endpoint**: `POST /join/:token`
- **Headers**: `Authorization: Bearer <token>`
- **Response**: `Group Object`

### Preview Invite Link (Public)
- **Endpoint**: `GET /join/:token`
- **Headers**: _None_
- **Response**:
  ```json
  {
    "group": {
      "id": 10,
      "name": "My Group",
      "description": "Cool people only",
      "icon": "",
      "is_public": false,
      "handle": null
    },
    "expires_at": "2025-01-01T01:00:00Z",
    "max_uses": null,
    "used_count": 0,
    "requires_auth": true
  }
  ```

### Leave Group
- **Endpoint**: `POST /groups/:id/leave`
- **Headers**: `Authorization: Bearer <token>`
- **Response**: `{"message": "Left group successfully"}`

### Get Group Members
- **Endpoint**: `GET /groups/:id/members`
- **Headers**: `Authorization: Bearer <token>`
- **Notes**: Private groups require membership.
- **Response**: `[ ...User Objects... ]`

### Get Group Messages
- **Endpoint**: `GET /groups/:id/messages?limit=50`
- **Headers**: `Authorization: Bearer <token>`
- **Ordering**: Newest-first by `id`. Use `next_cursor` (oldest ID in page) for pagination.
- **Response**:
  ```json
  {
    "messages": [ ...Message Objects... ],
    "count": 50,
    "next_cursor": 1200
  }
  ```

### Sync Messages (REST)
Incremental sync for background polling.
- **Endpoint**: `POST /messages/sync`
- **Headers**: `Authorization: Bearer <token>`
- **Body**:
  ```json
  {
    "limit": 100,
    "conversations": [
      { "conversation_id": "user_123", "last_message_id": 50 },
      { "conversation_id": "group_10", "last_message_id": 200 }
    ]
  }
  ```
- **Response**:
  ```json
  {
    "results": [
      {
        "conversation_id": "user_123",
        "messages": [ ...Message Objects... ],
        "has_more": false,
        "next_cursor": 75
      }
    ],
    "count": 1
  }
  ```

### Send Group Message (REST Fallback)
- **Endpoint**: `POST /groups/:id/messages`
- **Headers**: `Authorization: Bearer <token>`
- **Body**:
  ```json
  {
    "client_id": "uuid-v4",
    "content": "Hello group",
    "message_type": "text"
  }
  ```
- **Response**: `Message Object`

### Mark Group Read
- **Endpoint**: `POST /groups/:id/read`
- **Headers**: `Authorization: Bearer <token>`
- **Body**:
  ```json
  { "last_read_message_id": 1200 }
  ```
- **Notes**: Monotonic (never decreases). If `last_read_message_id` exceeds latest, it is clamped.
- **Response**:
  ```json
  {
    "ok": true,
    "last_read_message_id": 1200,
    "latest_group_message_id": 1210
  }
  ```

### Get Group Read State
- **Endpoint**: `GET /groups/:id/read-state`
- **Headers**: `Authorization: Bearer <token>`
- **Response**:
  ```json
  {
    "my_last_read_message_id": 1200,
    "members": [
      { "user_id": 1, "last_read_message_id": 1200 },
      { "user_id": 2, "last_read_message_id": 1180 }
    ]
  }
  ```

---

## WebSocket API (Real-Time)

**URL**: `ws://<host>:8080/ws`
**Auth**: Handled via Cookie or Query Param (depending on implementation, currently Cookie/Header based in middleware).

### Message Types (Client -> Server)

#### 1. Send Chat Message
```json
{
  "type": "chat",
  "client_id": "uuid-v4",
  "recipient_id": 123, // OR "group_id": 456
  "content": "Hello world",
  "message_type": "text"
}
```

#### 2. Typing Indicator
```json
{
  "type": "typing",
  "recipient_id": 123,
  "is_typing": true
}
```

#### 3. Mark as Read
```json
{
  "type": "read",
  "message_id": 999
}
```

#### 4. Mark Group Read
```json
{
  "type": "group_read",
  "group_id": 456,
  "last_read_message_id": 1200
}
```

#### 5. Sync Messages (Reconnect)
```json
{
  "type": "sync",
  "conversations": [
    {
      "conversation_id": "user_123", // or "group_456"
      "last_message_id": 50
    }
  ]
}
```

### Message Types (Server -> Client)

#### 1. New Message
```json
{
  "type": "message",
  "message": {
    "id": 100,
    "content": "Hello world",
    "sender": { ... },
    "created_at": "..."
  }
}
```

#### 2. Message ACK (Sent)
```json
{
  "type": "ack",
  "client_id": "uuid-v4",
  "server_id": 100,
  "status": "sent"
}
```

#### 3. Typing Indicator
```json
{
  "type": "typing",
  "sender_id": 123,
  "is_typing": true
}
```

#### 4. Batch Messages (On Reconnect)
When a user reconnects after being offline, the server sends pending messages in batches:
```json
{
  "type": "batch",
  "count": 25,
  "messages": [
    {
      "type": "message",
      "message": { ... }
    },
    {
      "type": "message",
      "message": { ... }
    }
  ]
}
```

#### 5. Group Read Update
```json
{
  "type": "group_read_update",
  "group_id": 456,
  "user_id": 3,
  "last_read_message_id": 1200
}
```

#### 6. DM Read Update
```json
{
  "type": "read_update",
  "conversation_id": "user_123",
  "user_id": 123,
  "last_read_message_id": 1200
}
```

---

## Ordering & Timestamps (Important)
- Message ordering is **newest-first by `id`** for both DM and group lists.
- Use `created_at_unix` (UTC seconds) for display ordering and convert to local time in the client.
- Cursor pagination uses `(created_at, id)` for deterministic conversation ordering.
**Note**: Batches are limited to 50 messages. If more messages are queued, multiple batches will be sent with a 100ms delay between them.

#### 5. Sync Response
```json
{
  "type": "sync_response",
  "conversation_id": "user_123",
  "messages": [ ...messages... ],
  "has_more": true,
  "next_cursor": 150
}
```

---

## Offline Message Queue & Delivery Guarantees

### How It Works

**Automatic Queueing**: When a message is sent to an offline user or delivery fails:
1. Message is stored in the `pending_messages` table
2. Includes retry metadata (attempts, next_retry timestamp)
3. Prioritized by message priority and creation time

**On Reconnection**:
1. User connects via WebSocket
2. Server immediately flushes all pending messages in batches (max 50 per batch)
3. Successfully delivered messages are removed from queue

**Retry Mechanism** (Background Worker):
- Runs every 5 seconds
- Checks for messages with `next_retry <= now()`
- Uses exponential backoff: 2s → 4s → 8s → 16s → 32s
- Max 5 retry attempts before entering long-term queue (1 hour delay)
- If user comes online, messages are delivered immediately

**Delivery Tracking**:
- Messages include `client_id` (UUID) for deduplication
- Duplicate sends return the same ACK (idempotent)
- Status progression: pending → sent → delivered → read

### Performance Characteristics

- **2G Network**: Batching reduces round-trips significantly
- **Intermittent Connectivity**: Messages queue during disconnects
- **No Message Loss**: All messages persisted until delivered
- **Efficient Bandwidth**: Batch envelope minimizes protocol overhead
