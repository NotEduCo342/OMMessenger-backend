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
- **Endpoint**: `GET /users/:username`
- **Headers**: `Authorization: Bearer <token>`
- **Response**: `{"user": { ... }}`

---

## Messages (REST)

### Get Messages
Fetch message history for a conversation.
- **Endpoint**: `GET /messages?recipient_id=123&limit=50`
- **Headers**: `Authorization: Bearer <token>`
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

### Create Group
- **Endpoint**: `POST /groups`
- **Headers**: `Authorization: Bearer <token>`
- **Body**:
  ```json
  {
    "name": "My Group",
    "description": "Cool people only"
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

### Leave Group
- **Endpoint**: `POST /groups/:id/leave`
- **Headers**: `Authorization: Bearer <token>`
- **Response**: `{"message": "Left group successfully"}`

### Get Group Members
- **Endpoint**: `GET /groups/:id/members`
- **Headers**: `Authorization: Bearer <token>`
- **Response**: `[ ...User Objects... ]`

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

#### 4. Sync Messages (Reconnect)
```json
{
  "type": "sync",
  "conversations": [
    {
      "conversation_id": "user_123",
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
