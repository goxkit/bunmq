# bunmq

[![Go Reference](https://pkg.go.dev/badge/github.com/goxkit/bunmq.svg)](https://pkg.go.dev/github.com/goxkit/bunmq)
[![Go Report Card](https://goreportcard.com/badge/github.com/goxkit/bunmq)](https://goreportcard.com/report/github.com/goxkit/bunmq)
[![ci](https://github.com/goxkit/bunmq/actions/workflows/ci.yml/badge.svg)](https://github.com/goxkit/bunmq/actions/workflows/ci.yml)
![Code Scanning](https://img.shields.io/github/alerts/goxkit/bunmq?label=security%20alerts)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A resilient and robust RabbitMQ library for Go applications, providing high-level abstractions for topology management, message publishing, and consuming with built-in retry mechanisms, dead letter queue (DLQ) support, and **automatic connection/channel recovery**.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Connection Management & Resilience](#connection-management--resilience)
  - [Automatic Reconnection](#automatic-reconnection)
  - [Connection Health Monitoring](#connection-health-monitoring)
  - [Reconnection Configuration](#reconnection-configuration)
- [Concepts](#concepts)
  - [Topology](#topology)
  - [Queues](#queues)
  - [Exchanges](#exchanges)
  - [Publishing](#publishing)
  - [Message Consumption](#message-consumption)
- [Error Handling and Resilience](#error-handling-and-resilience)
  - [Retry Strategy](#retry-strategy)
  - [Dead Letter Queues](#dead-letter-queues)
  - [Error Types](#error-types)
- [Observability](#observability)
  - [Distributed Tracing](#distributed-tracing)
  - [Logging](#logging)
- [Advanced Usage](#advanced-usage)
  - [Exchange Bindings](#exchange-bindings)
  - [Custom Message Metadata](#custom-message-metadata)
  - [Multiple Queue Definitions](#multiple-queue-definitions)
- [Configuration](#configuration)
  - [Connection Options](#connection-options)
  - [Queue Configuration Options](#queue-configuration-options)
- [Best Practices](#best-practices)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Automatic Connection Recovery**: Built-in connection and channel monitoring with automatic reconnection using `NotifyClose` and `NotifyCancel`
- **Exponential Backoff Strategy**: Configurable reconnection delays with exponential backoff to prevent overwhelming the broker
- **Health Monitoring**: Real-time connection and channel health checks with configurable callbacks
- **Topology Management**: Declarative configuration of exchanges, queues, and bindings
- **Retry Strategy**: Configurable retry mechanisms with exponential backoff
- **Dead Letter Queues**: Automatic routing of failed messages to DLQs
- **Distributed Tracing**: Built-in OpenTelemetry support for observability
- **Type-Safe Handlers**: Strongly typed message handlers with reflection-based dispatching
- **Graceful Shutdown**: Signal-based shutdown with proper resource cleanup
- **Resilient Publishing**: Publisher with deadline support and automatic reconnection

## Installation

```bash
go get github.com/goxkit/bunmq
```

## Quick Start

```go
package main

import (
    "context"
    "time"

    "github.com/goxkit/bunmq"
    "github.com/sirupsen/logrus"
)

type OrderCreated struct {
    ID     string  `json:"id"`
    Amount float64 `json:"amount"`
}

func main() {
    // Define queue with retry and DLQ
    queueDef := bunmq.
        NewQueue("my-queue").
        Durable(true).
        WithMaxLength(100_000).
        WithRetry(time.Second*10, 3).
        WithDQL().
        WithDLQMaxLength(10_000)

    // Create topology
    topology := bunmq.
        NewTopology("my-app", "amqp://guest:guest@localhost:5672/").
        Queue(queueDef).
        Exchange(bunmq.NewDirectExchange("orders")).
        QueueBinding(
            bunmq.NewQueueBinding().
                Queue("orders").
                Exchange("orders").
                RoutingKey("order.created"),
        )

    // Apply topology and get connection manager with automatic reconnection
    manager, err := topology.Apply()
    if err != nil {
        panic(err)
    }
    defer manager.Close()

    // Create dispatcher with resilient connection management
    dispatcher := bunmq.NewDispatcher(manager, []*bunmq.QueueDefinition{queueDef})
    
    err = dispatcher.RegisterByType("orders", &OrderCreated{}, func(ctx context.Context, msg any, metadata any) error {
        order := msg.(*OrderCreated)
        logrus.Infof("Processing order: %s, Amount: %.2f", order.ID, order.Amount)
        return nil
    })
    if err != nil {
        panic(err)
    }

    // Start consuming (blocks until signal)
    dispatcher.ConsumeBlocking()
}
```

## Connection Management & Resilience

### Automatic Reconnection

BunMQ provides robust connection management through the `ConnectionManager` that automatically handles connection and channel failures. The system uses RabbitMQ's `NotifyClose` and `NotifyCancel` mechanisms to detect failures and trigger reconnection with exponential backoff.

```go
// Connection manager is automatically created when applying topology
manager, err := topology.Apply()
if err != nil {
    panic(err)
}

// The manager automatically:
// 1. Monitors connection health using NotifyClose
// 2. Monitors channel health using NotifyCancel  
// 3. Reconnects with exponential backoff on failures
// 4. Reapplies topology after successful reconnection
// 5. Restarts consumers automatically
```

### Connection Health Monitoring

The connection manager continuously monitors both connection and channel health:

```go
// Check current health status
if manager.IsHealthy() {
    logrus.Info("Connection and channel are healthy")
}

// Set up reconnection callback for custom logic
manager.SetReconnectCallback(func(conn bunmq.RMQConnection, ch bunmq.AMQPChannel) {
    logrus.Info("Connection successfully reestablished")
    // Custom logic after reconnection (e.g., metrics, notifications)
})
```

### Reconnection Configuration

Customize the reconnection behavior with `ReconnectionConfig`:

```go
// Create connection manager with custom reconnection settings
config := bunmq.ReconnectionConfig{
    MaxAttempts:   10,              // Maximum reconnection attempts (0 = infinite)
    InitialDelay:  time.Second * 2, // Initial delay between attempts
    BackoffMax:    time.Minute * 5, // Maximum delay between attempts  
    BackoffFactor: 1.5,             // Exponential backoff factor
}

manager, err := bunmq.NewConnectionManager("my-app", connectionString, config)
```

**Default Configuration:**
- **MaxAttempts**: 0 (infinite attempts)
- **InitialDelay**: 2 seconds  
- **BackoffMax**: 5 minutes
- **BackoffFactor**: 1.5x exponential backoff

**Reconnection Strategy:**
1. **Channel Validation First**: Always validate channel health before connection
2. **Connection Validation**: Only check connection if channel validation fails
3. **Exponential Backoff**: Delays increase exponentially (2s → 3s → 4.5s → 6.75s → ...)
4. **Topology Reapplication**: Automatically redeclare exchanges, queues, and bindings
5. **Consumer Restart**: Automatically restart all registered consumers

## Concepts

### Topology

The topology defines the complete RabbitMQ infrastructure including exchanges, queues, and their bindings. This ensures your messaging infrastructure is properly configured before your application starts. With the new ConnectionManager, topology is automatically reapplied after reconnections.

```go
topology := bunmq.NewTopology("my-app", "amqp://guest:guest@localhost:5672/")

// Add exchanges
topology.Exchange(bunmq.NewDirectExchange("orders"))
topology.Exchange(bunmq.NewFanoutExchange("notifications"))

// Add queues with retry and DLQ
orderQueue := bunmq.NewQueue("orders").
    Durable(true).
    WithRetry(time.Second*30, 5).
    WithDQL()

topology.Queue(orderQueue)

// Bind queues to exchanges
topology.QueueBinding(
    bunmq.NewQueueBinding().
        Queue("orders").
        Exchange("orders").
        RoutingKey("order.created"),
)

// Apply the topology - returns ConnectionManager with automatic reconnection
manager, err := topology.Apply()
if err != nil {
    panic(err)
}
defer manager.Close()

// The topology will be automatically reapplied after any reconnection
```

### Queues

Queues support various configuration options including durability, TTL, retry mechanisms, and dead letter queues.

```go
// Basic queue
basicQueue := bunmq.NewQueue("basic-queue")

// Durable queue with TTL
ttlQueue := bunmq.NewQueue("ttl-queue").
    Durable(true).
    WithTTL(time.Hour * 24)

// Queue with retry and DLQ
resilientQueue := bunmq.NewQueue("resilient-queue").
    Durable(true).
    WithRetry(time.Second*30, 3).  // 30s delay, 3 retries
    WithDQL()                      // Enable dead letter queue

// Exclusive queue (deleted when connection closes)
exclusiveQueue := bunmq.NewQueue("exclusive-queue").
    Exclusive(true)
```

### Exchanges

Multiple exchange types are supported with a fluent API:

```go
// Direct exchange (point-to-point routing)
directExch := bunmq.NewDirectExchange("orders")

// Fanout exchange (broadcast to all bound queues)  
fanoutExch := bunmq.NewFanoutExchange("notifications")

// Multiple exchanges at once
exchanges := bunmq.NewDirectExchanges([]string{"orders", "payments", "inventory"})
topology.Exchanges(exchanges)
```

### Publishing

The publisher provides methods for sending messages with optional routing and deadline support. It can work with both direct channels and the resilient ConnectionManager:

```go
// Get channel from connection manager (automatically handles reconnection)
channel, err := manager.GetChannel()
if err != nil {
    return err
}

publisher := bunmq.NewPublisher("my-app", channel)

// Simple publish to exchange
exchange := "orders"
routingKey := "order.created"
message := OrderCreated{ID: "123", Amount: 99.99}

err := publisher.Publish(ctx, &exchange, nil, &routingKey, message)

// Publish with deadline (1 second timeout)
err = publisher.PublishDeadline(ctx, &exchange, nil, &routingKey, message)

// The publisher will automatically use the reconnected channel if a reconnection occurs
```

### Message Consumption

The dispatcher handles message routing to type-safe handlers with automatic retry, error handling, and **automatic consumer restart** on connection failures:

```go
// Create dispatcher with ConnectionManager for automatic reconnection
dispatcher := bunmq.NewDispatcher(manager, []*bunmq.QueueDefinition{queueDef})

// Register typed handler
err := dispatcher.RegisterByType("orders", &OrderCreated{}, func(ctx context.Context, msg any, metadata *bunmq.DeliveryMetadata) error {
    order := msg.(*OrderCreated)
    
    // Process the order
    if err := processOrder(order); err != nil {
        // Return RetryableError for transient failures
        if isTransientError(err) {
            return bunmq.RetryableError
        }
        // Other errors will send message to DLQ
        return err
    }
    
    return nil
})

// Start consuming (blocks until SIGTERM/SIGINT)
// Consumers are automatically restarted after reconnections
dispatcher.ConsumeBlocking()
```

**Key Features:**
- **Automatic Consumer Restart**: All registered consumers are automatically restarted after connection recovery
- **Seamless Failover**: No message loss during reconnection (messages remain in queues)
- **Preserved Handler State**: All registered handlers and their configurations are maintained across reconnections

## Error Handling and Resilience

### Connection and Channel Failures

BunMQ automatically handles connection and channel failures through the ConnectionManager:

```go
// Connection failures are automatically detected and handled
// No manual intervention required

// Optional: Set up monitoring for reconnection events
manager.SetReconnectCallback(func(conn bunmq.RMQConnection, ch bunmq.AMQPChannel) {
    logrus.Info("Connection reestablished after failure")
    
    // Send notification to monitoring system
    metrics.IncrementCounter("rabbitmq.reconnections")
})

// Check connection health programmatically
if !manager.IsHealthy() {
    logrus.Warn("RabbitMQ connection is currently unhealthy")
}
```

**Failure Scenarios Handled:**
- **Network partitions**: Automatic reconnection with exponential backoff
- **RabbitMQ server restarts**: Topology and consumers are automatically restored
- **Channel closures**: New channels are created and configured
- **Authentication failures**: Logged with appropriate error messages
- **Resource constraints**: Backoff prevents overwhelming the server

### Retry Strategy

Messages that fail processing can be automatically retried based on the queue configuration:

```go
queueDef := bunmq.NewQueue("orders").
    WithRetry(time.Second*30, 3)  // 30 second delay, 3 retries max
```

When a handler returns `bunmq.RetryableError`, the message will be:
1. Sent to the retry queue with TTL
2. After TTL expires, requeued to the original queue
3. Processed again up to the retry limit
4. Sent to DLQ if retries are exhausted

### Dead Letter Queues

Failed messages are automatically routed to dead letter queues for manual inspection:

```go
queueDef := bunmq.NewQueue("orders").
    WithDQL()  // Creates "orders-dlq" automatically
```

Messages are sent to the DLQ when:
- Handler returns a non-retryable error
- Retry limit is exceeded
- Message processing fails permanently

### Error Types

```go
// Transient errors that should be retried
if temporaryFailure {
    return bunmq.RetryableError
}

// Permanent errors that go directly to DLQ
return fmt.Errorf("invalid message format: %w", err)
```

## Observability

### Distributed Tracing

Built-in OpenTelemetry support for end-to-end tracing across services:

```go
// Trace context is automatically propagated through message headers
// and extracted in consumer handlers

func handleOrder(ctx context.Context, msg any, metadata any) error {
    // ctx contains the trace from the publisher
    span := trace.SpanFromContext(ctx)
    span.SetAttributes(attribute.String("order.id", order.ID))
    
    return processOrder(ctx, order)
}
```

### Logging

Structured logging is provided throughout the library using logrus:

```go
// Set log level for bunmq operations
logrus.SetLevel(logrus.DebugLevel)
```

## Advanced Usage

### Exchange Bindings

Create complex routing topologies by binding exchanges to other exchanges:

```go
// Route messages from direct to fanout exchange
exchangeBinding := bunmq.NewExchangeBiding().
    Source("orders").
    Destination("notifications").
    RoutingKey("order.created")

topology.ExchangeBinding(exchangeBinding)
```

### Custom Message Metadata

Access message metadata in handlers for advanced processing:

```go
func handleMessage(ctx context.Context, msg any, metadata any) error {
    meta := metadata.(*bunmq.DeliveryMetadata)
    
    logrus.Infof("Message ID: %s, Retry Count: %d", 
                 meta.MessageID, meta.XCount)
    
    return nil
}
```

### Multiple Queue Definitions

Manage multiple queues with different configurations:

```go
orderQueue := bunmq.NewQueue("orders").WithRetry(time.Second*30, 3)
paymentQueue := bunmq.NewQueue("payments").WithDQL()
notificationQueue := bunmq.NewQueue("notifications").Durable(false)

topology.Queues([]*bunmq.QueueDefinition{
    orderQueue, paymentQueue, notificationQueue,
})
```

## Configuration

### Connection Options

```go
// Basic connection with automatic reconnection
topology := bunmq.NewTopology("my-app", "amqp://guest:guest@localhost:5672/")

// With TLS (automatic reconnection included)
topology := bunmq.NewTopology("my-app", "amqps://user:pass@rabbitmq.example.com:5671/")

// With vhost (automatic reconnection included)
topology := bunmq.NewTopology("my-app", "amqp://user:pass@localhost:5672/my-vhost")

// Direct ConnectionManager creation with custom reconnection config
config := bunmq.ReconnectionConfig{
    MaxAttempts:   5,               // Limit reconnection attempts
    InitialDelay:  time.Second * 1, // Faster initial reconnection
    BackoffMax:    time.Minute * 2, // Lower maximum delay
    BackoffFactor: 2.0,             // Faster exponential growth
}

manager, err := bunmq.NewConnectionManager("my-app", connectionString, config)
```

### Queue Configuration Options

| Method | Description | Default |
|--------|-------------|---------|
| `Durable(bool)` | Survive broker restarts | `true` |
| `Delete(bool)` | Auto-delete when unused | `false` |
| `Exclusive(bool)` | Used by one connection only | `false` |
| `WithMaxLength(max)` | Max number of messages retained in the queue | Not set |
| `WithTTL(duration)` | Message time-to-live | Not set |
| `WithRetry(ttl, retries)` | Retry configuration | Disabled |
| `WithDQL()` | Enable dead letter queue | Disabled |
| `WithDQLWithDLQMaxLength(max)` | Max number of messages retained in the DLQ  | Not set |

## Best Practices

1. **Always use durable queues** for important messages
2. **Enable DLQs** for production workloads to handle permanent failures
3. **Configure appropriate retry policies** based on your use case
4. **Use the ConnectionManager** for automatic reconnection in production environments
5. **Monitor reconnection events** using callbacks for operational visibility
6. **Configure reasonable reconnection limits** to prevent infinite retry loops in persistent failures
7. **Use typed message handlers** to ensure type safety
8. **Implement proper error classification** (retryable vs permanent)
9. **Monitor DLQs** and set up alerts for failed messages
10. **Use structured logging** for better observability
11. **Test your topology** in development environments first
12. **Test reconnection scenarios** by simulating network failures and RabbitMQ restarts

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.