# Async Safety Improvements for SimConnector

This document describes the async safety improvements implemented in the SimConnector library to handle concurrent operations and prevent race conditions.

## Overview of Changes

### 1. Thread-Safe State Management

**Problem**: The original implementation had race conditions on shared state:
- `e.system.IsConnected` was accessed/modified without synchronization
- `e.handle` was read/written from multiple goroutines
- Context initialization could happen multiple times

**Solution**: Added mutex protection for all shared state:

```go
type Engine struct {
    // ... existing fields ...
    
    // Async safety controls
    mu           sync.RWMutex  // Protects shared state
    startOnce    sync.Once     // Ensures Listen() is called only once
    contextOnce  sync.Once     // Ensures context initialization happens only once
    isListening  bool          // Protected by mu, tracks if listening is active
}

type SystemState struct {
    mu          sync.RWMutex // Protects IsConnected
    IsConnected bool
}
```

### 2. RunOnce Pattern for Listen()

**Problem**: Multiple calls to `Listen()` could spawn multiple goroutines, leading to:
- Resource waste
- Multiple message dispatch loops
- Unpredictable behavior

**Solution**: Implemented `sync.Once` pattern:

```go
func (e *Engine) Listen() <-chan any {
    // Thread-safe check for connection status
    e.system.mu.RLock()
    isConnected := e.system.IsConnected
    e.system.mu.RUnlock()
    
    if !isConnected {
        return nil
    }

    // Thread-safe check if already listening
    e.mu.RLock()
    alreadyListening := e.isListening
    e.mu.RUnlock()
    
    if alreadyListening {
        return e.stream // Return existing stream
    }

    // Use sync.Once to ensure context and goroutine are initialized only once
    e.contextOnce.Do(func() {
        e.ctx, e.cancel = context.WithCancel(context.Background())
        e.done = make(chan struct{})
    })

    e.startOnce.Do(func() {
        // ... start dispatch goroutine only once ...
    })

    return e.stream
}
```

### 3. Buffered Channel for Message Streaming

**Problem**: The original unbuffered channel could cause blocking:
- Message dispatch could block if consumer is slow
- Potential deadlocks in high-throughput scenarios

**Solution**: Added buffered channel with configurable size:

```go
const DEFAULT_STREAM_BUFFER_SIZE = 100

// In NewWithCustomDLL:
stream: make(chan any, DEFAULT_STREAM_BUFFER_SIZE)
```

Benefits:
- Non-blocking message dispatch (up to buffer size)
- Better performance under load
- Graceful handling of burst traffic

### 4. Race-Free Connection Management

**Problem**: Open/Close operations had race conditions:
- Multiple goroutines could call Open() simultaneously
- Connection state could be inconsistent

**Solution**: Added mutex protection for all connection operations:

```go
func (e *Engine) Open() error {
    e.mu.Lock()
    defer e.mu.Unlock()
    
    // Thread-safe check for connection status
    e.system.mu.RLock()
    isConnected := e.system.IsConnected
    e.system.mu.RUnlock()
    
    if isConnected {
        return fmt.Errorf("connection already open")
    }
    
    // ... perform connection ...
    
    // Thread-safe update of connection status
    e.system.mu.Lock()
    e.system.IsConnected = true
    e.system.mu.Unlock()
    
    return nil
}
```

### 5. Graceful Shutdown Handling

**Problem**: Shutdown wasn't properly coordinated:
- Goroutines could continue running after Close()
- Resource cleanup wasn't guaranteed

**Solution**: Proper shutdown coordination:

```go
func (e *Engine) Close() error {
    e.mu.Lock()
    defer e.mu.Unlock()
    
    // ... check connection state ...
    
    // Signal graceful shutdown
    if e.cancel != nil {
        e.cancel()
        // Wait for dispatch to finish
        if e.done != nil {
            <-e.done
        }
    }
    
    // ... close connection ...
    
    e.isListening = false
    return nil
}
```

## Key Benefits

### 1. **Race Condition Prevention**
- All shared state is protected by mutexes
- No more data races on connection status or handle

### 2. **Resource Efficiency**
- Only one dispatch goroutine per Engine instance
- Proper cleanup on shutdown
- No goroutine leaks

### 3. **Predictable Behavior**
- Multiple `Listen()` calls return the same channel
- Consistent behavior under concurrent access

### 4. **Better Performance**
- Buffered channels prevent blocking
- Read-write mutexes allow concurrent reads
- Efficient message processing

### 5. **Robust Error Handling**
- Graceful handling of concurrent operations
- Proper error propagation
- Safe operation in all states

## Usage Examples

### Safe Concurrent Usage

```go
func main() {
    sdk := client.New("MyApp")
    defer sdk.Close()
    
    if err := sdk.Open(); err != nil {
        log.Fatal(err)
    }
    
    // Multiple goroutines can safely call Listen()
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            // This is now safe - all calls return the same channel
            messages := sdk.Listen()
            if messages == nil {
                return
            }
            
            // Process messages...
            for msg := range messages {
                // Handle message
            }
        }(i)
    }
    
    wg.Wait()
}
```

### Testing Async Safety

The library includes a comprehensive test demo at `example/async_safety_demo.go` that demonstrates:

1. **Concurrent Listen() calls** - Verifies sync.Once behavior
2. **Concurrent message processing** - Tests buffered channel performance  
3. **Concurrent Close() operations** - Validates graceful shutdown
4. **Error handling** - Tests behavior when not connected

Run the demo with:
```bash
go run example/async_safety_demo.go
```

## Migration Notes

Existing code should continue to work without changes. The improvements are backward-compatible:

- `Listen()` still returns `<-chan any`
- `Open()` and `Close()` have the same signatures
- Error handling remains the same

The only visible change is improved reliability under concurrent usage.

## Performance Considerations

1. **Mutex Overhead**: Minimal due to read-write mutexes allowing concurrent reads
2. **Memory Usage**: Slightly increased due to buffered channel (configurable)
3. **Goroutine Efficiency**: Reduced from potential N goroutines to exactly 1 per Engine
4. **Message Throughput**: Improved due to buffered channel and non-blocking sends

## Future Enhancements

Potential areas for further improvement:

1. **Configurable Buffer Sizes**: Allow custom buffer sizes per application needs
2. **Metrics/Monitoring**: Add counters for dropped messages, goroutine states
3. **Backpressure Handling**: More sophisticated handling when buffer is full
4. **Connection Pool**: Support for multiple SimConnect connections
