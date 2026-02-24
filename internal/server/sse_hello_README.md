# SSE Hello World Implementation

## Overview

This is a simple "Hello World" implementation of Server-Sent Events (SSE) in Go, demonstrating the basic concepts of SSE streaming.

## Files

- `sse_hello.go` - Implementation of the SSE hello endpoint
- `sse_hello_test.go` - Comprehensive tests for SSE functionality
- `test_helpers.go` - Shared test utilities

## Features

### Basic SSE Streaming
- Sends configurable number of "hello" events
- Each event contains:
  - Message: "Hello, World!"
  - Count: Current event number
  - Total: Total number of events
  - Time: RFC3339 formatted timestamp

### Query Parameters
- `count` (optional): Number of hello events to send (default: 1)

### Response Format

The endpoint follows the SSE protocol:

```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

event: hello
data: {"message":"Hello, World!","count":1,"total":1,"time":"2026-01-26T..."}

event: complete
data: {"message":"SSE stream complete","total":1}
```

## Usage Examples

### Single Event (Default)
```bash
curl http://localhost:8080/sse/hello
```

### Multiple Events
```bash
curl http://localhost:8080/sse/hello?count=5
```

### JavaScript Client Example
```javascript
const eventSource = new EventSource('/sse/hello?count=3');

eventSource.addEventListener('hello', (e) => {
  const data = JSON.parse(e.data);
  console.log(`Hello ${data.count}/${data.total}: ${data.message}`);
});

eventSource.addEventListener('complete', (e) => {
  const data = JSON.parse(e.data);
  console.log(data.message);
  eventSource.close();
});

eventSource.onerror = (error) => {
  console.error('SSE error:', error);
  eventSource.close();
};
```

## Testing

### Run All SSE Tests
```bash
go test -v ./internal/server -run TestSSEHelloWorld
```

### Test Coverage
```bash
go test -cover ./internal/server -run TestSSEHelloWorld
```

### Test Details

1. **TestSSEHelloWorld** - Verifies basic SSE functionality
   - Correct SSE headers (Content-Type, Cache-Control, Connection)
   - Event format compliance
   - Message content accuracy

2. **TestSSEHelloWorldMultipleEvents** - Tests streaming multiple events
   - Configurable event count via query parameter
   - Correct event sequencing

3. **TestSSEHelloWorldTiming** - Verifies real-time streaming
   - Events arrive promptly
   - No buffering delays

## Implementation Details

### SSE Event Structure

Each SSE event follows this format:
```
event: <event-type>
data: <json-payload>
<blank line>
```

### Error Handling

- JSON marshaling errors: Logged and skipped (stream continues)
- Streaming not supported: Returns HTTP 500 error
- Invalid count parameter: Defaults to 1

### Performance Considerations

- Small delay (10ms) between events for realistic streaming
- Immediate flushing after each event (no buffering)
- Lightweight JSON payloads

## Integration with AgentServer

The `handleSSEHello` method is a receiver on `AgentServer`, allowing it to:
- Access server configuration
- Share resources with other handlers
- Be easily registered as an HTTP route

Example registration:
```go
http.HandleFunc("/sse/hello", server.handleSSEHello)
```

## Design Decisions

### TDD Approach
This implementation followed strict Test-Driven Development:
1. **RED**: Wrote failing tests first
2. **GREEN**: Implemented minimal code to pass tests
3. **REFACTOR**: Enhanced with documentation and helper functions

### Separation of Concerns
- `sendSSEEvent()` - Reusable helper for formatting SSE events
- `handleSSEHello()` - Business logic for hello endpoint
- Clean separation allows for future SSE endpoints

### Type Safety
- Uses Go's type system for compile-time safety
- JSON marshaling with proper error handling
- Interface-based flusher detection

## Future Enhancements

Potential improvements:
- [ ] Add event IDs for resume capability
- [ ] Support retry field for client reconnection
- [ ] Add custom message via query parameter
- [ ] Implement server-side heartbeat/keepalive
- [ ] Add metrics/monitoring for stream health

## References

- [MDN: Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events)
- [W3C: Server-Sent Events Spec](https://html.spec.whatwg.org/multipage/server-sent-events.html)
- [Go net/http Package](https://pkg.go.dev/net/http)
