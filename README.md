# bunmq

[![Go Reference](https://pkg.go.dev/badge/github.com/goxkit/bunmq.svg)](https://pkg.go.dev/github.com/goxkit/bunmq)
[![Go Report Card](https://goreportcard.com/badge/github.com/goxkit/bunmq)](https://goreportcard.com/report/github.com/goxkit/bunmq)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A resilient and robust RabbitMQ library for Go applications, providing high-level abstractions for topology management, message publishing, and consuming with built-in retry mechanisms and dead letter queue (DLQ) support.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
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

    // Apply topology and get connection/channel
    manager, err := topology.Apply()
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    // Create dispatcher and register handler
    dispatcher := bunmq.NewDispatcher(manager, []*bunmq.QueueDefinition{queueDef})
    
    err = dispatcher.RegisterByType("orders", OrderCreated{}, func(ctx context.Context, msg any, metadata any) error {
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

## Concepts

### Topology

The topology defines the complete RabbitMQ infrastructure including exchanges, queues, and their bindings. This ensures your messaging infrastructure is properly configured before your application starts.

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

// Apply the topology
manager, err := topology.Apply()
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

The publisher provides methods for sending messages with optional routing and deadline support:

```go
publisher := bunmq.NewPublisher("my-app", channel)

// Simple publish to exchange
exchange := "orders"
routingKey := "order.created"
message := OrderCreated{ID: "123", Amount: 99.99}

err := publisher.Publish(ctx, &exchange, nil, &routingKey, message)

// Publish with deadline (1 second timeout)
err = publisher.PublishDeadline(ctx, &exchange, nil, &routingKey, message)
```

### Message Consumption

The dispatcher handles message routing to type-safe handlers with automatic retry and error handling:

```go
dispatcher := bunmq.NewDispatcher(manager, []*bunmq.QueueDefinition{queueDef})

// Register typed handler
err := dispatcher.RegisterByType("orders", OrderCreated{}, func(ctx context.Context, msg any, metadata any) error {
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
dispatcher.ConsumeBlocking()
```

## Error Handling and Resilience

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
// Basic connection
topology := bunmq.NewTopology("my-app", "amqp://guest:guest@localhost:5672/")

// With TLS
topology := bunmq.NewTopology("my-app", "amqps://user:pass@rabbitmq.example.com:5671/")

// With vhost
topology := bunmq.NewTopology("my-app", "amqp://user:pass@localhost:5672/my-vhost")
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
4. **Use typed message handlers** to ensure type safety
5. **Implement proper error classification** (retryable vs permanent)
6. **Monitor DLQs** and set up alerts for failed messages
7. **Use structured logging** for better observability
8. **Test your topology** in development environments first

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.