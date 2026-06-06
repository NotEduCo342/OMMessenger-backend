# OM Messenger Backend Analysis

This document provides a summary of the architecture, strengths, weaknesses, and bugs found in the OM Messenger backend, with a specific focus on the real-time WebSocket implementation and messaging reliability.

## Architecture & Strengths

- **Solid Foundation:** The backend is built on Go and Fiber, with PostgreSQL and GORM for persistence, which offers high performance and a well-structured ORM layer.
- **Caching Layer:** Redis is correctly integrated to cache conversations, group conversations, and unread counts, which helps significantly with read performance.
- **WebSocket Hub:** A dedicated Hub correctly manages user connections, keeping track of connected clients with an active ping/pong health checker. Compression is dynamically enabled for clients that support gzip, reducing bandwidth overhead.
- **Pending Message Queue:** A `PendingMessageRepository` is used to persist messages for offline users. Background workers attempt delivery using exponential backoff, which is a great approach to handling transient network failures.

## Weaknesses & Bugs

### 1. WebSocket Reliability and Messaging Order
**Critical Bug in Message Acknowledgment:** The current implementation assumes that a successful `Conn.WriteMessage()` over a WebSocket connection guarantees delivery.
- In `FlushPendingMessages`, the hub pulls pending messages, writes them to the connection, and **immediately deletes** them from the `PendingMessageRepository`.
- In `retryWorker`, the same logic applies: if `WriteMessage` doesn't return an error, the message is **immediately deleted**.
**Consequence:** If the client's network drops exactly after the server writes to the socket but before the packet reaches the client, the message is permanently lost from the queue. It will never be delivered to the client again unless the client performs a manual sync.
**Fix:** The backend must rely on application-level acknowledgments (`MessageAck`). Messages should remain in the `PendingMessageRepository` until the client explicitly sends a `MessageAck` confirming receipt.

### 2. General Code Quality
- **Unused Code:** `decompressData` in `hub.go` is implemented but never used.
- **Error Handling:** Some database and redis errors are logged but not strictly handled, though the application handles HTTP responses reasonably well.
- **Queueing Ephemeral Messages:** The hub attempts to exclude ephemeral messages (like typing, ping) from the offline queue, but its detection relies on casting `data` to `map[string]interface{}`. When strong-typed structs are passed to `SendToUser` (such as `models.MessageResponse`), the `dataMap, ok := data.(map[string]interface{})` cast fails, and the ephemeral check is bypassed. Since most data sent to `SendToUserWithID` is serialized differently, the priority system isn't robustly typing-aware.

### 3. Concurrency
- `clientsMux` is appropriately used in the Hub to protect the connections map.
- Background goroutines handle caching and offline user status asynchronously upon disconnects, avoiding blocking the socket closure.

## Proposed Fixes
1. Stop deleting messages from the `PendingMessageRepository` inside `FlushPendingMessages` and `retryWorker`.
2. Update the `Process` method for `MessageAck` to delete the pending message from the database.
3. Remove unused `decompressData` function.
