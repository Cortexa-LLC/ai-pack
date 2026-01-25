# Log Streaming API

The A2A agent server provides realtime log streaming endpoints for monitoring and debugging.

## Endpoints

### 1. Stream Logs (SSE)

**Endpoint:** `GET /logs/stream`

Streams logs in realtime using Server-Sent Events (SSE).

**Response Format:**
```
event: connected
data: {"message":"Log stream connected"}

event: log
data: {"timestamp":"2026-01-24T22:10:00Z","level":"INFO","message":"task_spawned","attrs":{"task_id":"task-123","role":"engineer"}}

event: log
data: {"timestamp":"2026-01-24T22:10:01Z","level":"INFO","message":"api_call","attrs":{"task_id":"task-123","tokens":1500}}
```

**Example (curl):**
```bash
curl -N http://localhost:8080/logs/stream
```

**Example (JavaScript):**
```javascript
const eventSource = new EventSource('http://localhost:8080/logs/stream');

eventSource.addEventListener('log', (event) => {
  const logEntry = JSON.parse(event.data);
  console.log(`[${logEntry.level}] ${logEntry.message}`, logEntry.attrs);
});

eventSource.addEventListener('connected', (event) => {
  console.log('Connected to log stream');
});
```

### 2. Recent Logs (JSON)

**Endpoint:** `GET /logs/recent?limit=N`

Returns the most recent N log entries as JSON (default: 100, max: 1000).

**Query Parameters:**
- `limit` (optional): Number of log entries to return (default: 100, max: 1000)

**Response Format:**
```json
{
  "logs": [
    {
      "timestamp": "2026-01-24T22:10:00Z",
      "level": "INFO",
      "message": "task_spawned",
      "attrs": {
        "task_id": "task-123",
        "role": "engineer",
        "status": "queued"
      }
    }
  ],
  "count": 100,
  "limit": 100
}
```

**Example:**
```bash
# Get last 50 log entries
curl http://localhost:8080/logs/recent?limit=50 | jq '.'

# Get last 10 entries, filter for errors only
curl http://localhost:8080/logs/recent?limit=100 | jq '.logs[] | select(.level == "ERROR")'
```

## Log Entry Format

Each log entry contains:

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | string | ISO 8601 timestamp |
| `level` | string | Log level: INFO, WARN, ERROR |
| `message` | string | Log message/event type |
| `attrs` | object | Structured attributes (task_id, role, etc.) |

## Common Log Messages

| Message | Description | Attributes |
|---------|-------------|------------|
| `server_starting` | Server startup | `version`, `max_concurrent`, `model` |
| `server_ready` | Server ready | `address`, `max_concurrent` |
| `task_spawned` | New task created | `task_id`, `role`, `task`, `status` |
| `task_started` | Task execution started | `task_id`, `role`, `status` |
| `task_completed` | Task completed | `task_id`, `role`, `status`, `duration_ms` |
| `task_failed` | Task failed | `task_id`, `role`, `error`, `duration_ms` |
| `api_call` | Anthropic API call | `task_id`, `model`, `tokens` |
| `http_request` | HTTP request | `method`, `path`, `status_code`, `duration_ms` |

## Buffer Configuration

The server maintains a ring buffer of the last **1000 log entries** in memory.

- Log entries older than the buffer size are automatically discarded
- The buffer is stored in memory only (not persisted)
- On server restart, the buffer is cleared

## Use Cases

### 1. Monitor Task Execution

```bash
curl -N http://localhost:8080/logs/stream | grep "task_"
```

### 2. Debug API Call Failures

```bash
curl http://localhost:8080/logs/recent?limit=500 | jq '.logs[] | select(.message == "api_call" or .level == "ERROR")'
```

### 3. Track Server Performance

```bash
curl http://localhost:8080/logs/recent?limit=100 | jq '.logs[] | select(.attrs.duration_ms != null) | {message, duration_ms: .attrs.duration_ms}'
```

### 4. Build a Dashboard

Use the SSE stream to build a realtime monitoring dashboard:

```html
<!DOCTYPE html>
<html>
<head>
  <title>A2A Server Logs</title>
</head>
<body>
  <h1>Realtime Logs</h1>
  <div id="logs"></div>

  <script>
    const eventSource = new EventSource('http://localhost:8080/logs/stream');
    const logsDiv = document.getElementById('logs');

    eventSource.addEventListener('log', (event) => {
      const log = JSON.parse(event.data);
      const entry = document.createElement('div');
      entry.className = `log-${log.level.toLowerCase()}`;
      entry.textContent = `[${log.timestamp}] [${log.level}] ${log.message}`;
      logsDiv.prepend(entry);

      // Keep only last 50 entries
      while (logsDiv.children.length > 50) {
        logsDiv.removeChild(logsDiv.lastChild);
      }
    });
  </script>
</body>
</html>
```

## Notes

- Log streaming is **read-only** - you cannot modify log levels or filters via the API
- The stream automatically reconnects if the connection is lost (SSE feature)
- For production deployments, consider using structured logging aggregators (ELK, Splunk, etc.)
- The buffer size (1000 entries) is currently hardcoded but can be made configurable
