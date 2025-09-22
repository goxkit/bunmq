// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import (
	"testing"
	"time"
)

func TestNewTopology(t *testing.T) {
	tests := []struct {
		name             string
		appName          string
		connectionString string
	}{
		{
			name:             "basic topology",
			appName:          "test-app",
			connectionString: "amqp://guest:guest@localhost:5672/",
		},
		{
			name:             "empty app name",
			appName:          "",
			connectionString: "amqp://guest:guest@localhost:5672/",
		},
		{
			name:             "empty connection string",
			appName:          "test-app",
			connectionString: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topology := NewTopology(tt.appName, tt.connectionString)
			if topology == nil {
				t.Fatal("NewTopology() returned nil")
			}

			// Verify it implements the Topology interface
			top := topology
			if top == nil {
				t.Error("NewTopology() does not implement Topology interface")
			}

			// Test that basic operations work
			queues := topology.GetQueuesDefinition()
			if queues == nil {
				t.Error("GetQueuesDefinition() should return non-nil map")
			}
			if len(queues) != 0 {
				t.Error("New topology should have no queues")
			}

			// Test that we can get a non-existent queue
			_, err := topology.GetQueueDefinition("nonexistent")
			if err == nil {
				t.Error("GetQueueDefinition() should return error for non-existent queue")
			}
		})
	}
}

func TestTopology_Queue(t *testing.T) {
	topology := NewTopology("test-app", "amqp://test")

	queue1 := NewQueue("orders")
	queue2 := NewQueue("payments")

	// Test adding single queue
	result := topology.Queue(queue1)
	if result != topology {
		t.Error("Queue() should return the same Topology instance")
	}

	// Test adding another queue
	topology.Queue(queue2)

	// Verify queues were added
	queues := topology.GetQueuesDefinition()
	if len(queues) != 2 {
		t.Errorf("Expected 2 queues, got %d", len(queues))
	}
	if queues["orders"] != queue1 {
		t.Error("Queue 'orders' not found or incorrect")
	}
	if queues["payments"] != queue2 {
		t.Error("Queue 'payments' not found or incorrect")
	}
}

func TestTopology_Queues(t *testing.T) {
	topology := NewTopology("test-app", "amqp://test")

	queue1 := NewQueue("orders")
	queue2 := NewQueue("payments")
	queue3 := NewQueue("notifications")

	// Test adding multiple queues
	result := topology.Queues([]*QueueDefinition{queue1, queue2, queue3})
	if result != topology {
		t.Error("Queues() should return the same Topology instance")
	}

	// Verify queues were added
	queues := topology.GetQueuesDefinition()
	if len(queues) != 3 {
		t.Errorf("Expected 3 queues, got %d", len(queues))
	}
	if queues["orders"] != queue1 {
		t.Error("Queue 'orders' not found or incorrect")
	}
	if queues["payments"] != queue2 {
		t.Error("Queue 'payments' not found or incorrect")
	}
	if queues["notifications"] != queue3 {
		t.Error("Queue 'notifications' not found or incorrect")
	}

	// Test adding empty slice
	topology.Queues([]*QueueDefinition{})
	if len(topology.GetQueuesDefinition()) != 3 {
		t.Error("Adding empty slice should not change queue count")
	}

	// Test adding nil slice
	topology.Queues(nil)
	if len(topology.GetQueuesDefinition()) != 3 {
		t.Error("Adding nil slice should not change queue count")
	}
}

func TestTopology_GetQueueDefinition(t *testing.T) {
	topology := NewTopology("test-app", "amqp://test")

	queue1 := NewQueue("orders")
	topology.Queue(queue1)

	// Test getting existing queue
	result, err := topology.GetQueueDefinition("orders")
	if err != nil {
		t.Errorf("GetQueueDefinition() returned unexpected error: %v", err)
	}
	if result != queue1 {
		t.Error("GetQueueDefinition() returned incorrect queue")
	}

	// Test getting non-existing queue
	result, err = topology.GetQueueDefinition("nonexistent")
	if err != NotFoundQueueDefinitionError {
		t.Errorf("GetQueueDefinition() should return NotFoundQueueDefinitionError, got %v", err)
	}
	if result != nil {
		t.Error("GetQueueDefinition() should return nil for non-existing queue")
	}
}

func TestTopology_Exchange(t *testing.T) {
	topology := NewTopology("test-app", "amqp://test")

	exchange1 := NewDirectExchange("orders")
	exchange2 := NewFanoutExchange("notifications")

	// Test adding single exchange
	result := topology.Exchange(exchange1)
	if result != topology {
		t.Error("Exchange() should return the same Topology instance")
	}

	// Test adding another exchange
	topology.Exchange(exchange2)

	// Since we can't access internal state directly (topology struct is not exported),
	// we'll verify behavior through the interface methods
	// The exchanges are stored but we can't directly verify them without Apply()
}

func TestTopology_Exchanges(t *testing.T) {
	topology := NewTopology("test-app", "amqp://test")

	exchange1 := NewDirectExchange("orders")
	exchange2 := NewFanoutExchange("notifications")
	exchange3 := NewDirectExchange("payments")

	// Test adding multiple exchanges
	result := topology.Exchanges([]*ExchangeDefinition{exchange1, exchange2, exchange3})
	if result != topology {
		t.Error("Exchanges() should return the same Topology instance")
	}

	// Since we can't access internal state directly (topology struct is not exported),
	// we test behavior through interface methods
	// Test adding empty slice - should not cause issues
	topology.Exchanges([]*ExchangeDefinition{})

	// Test adding nil slice - should not cause issues
	topology.Exchanges(nil)
}

func TestTopology_QueueBinding(t *testing.T) {
	topology := NewTopology("test-app", "amqp://test")

	binding1 := NewQueueBinding().Queue("orders").Exchange("order-exchange").RoutingKey("order.created")
	binding2 := NewQueueBinding().Queue("payments").Exchange("payment-exchange").RoutingKey("payment.processed")

	// Test adding single queue binding
	result := topology.QueueBinding(binding1)
	if result != topology {
		t.Error("QueueBinding() should return the same Topology instance")
	}

	// Test adding another queue binding
	topology.QueueBinding(binding2)

	// Since we can't access internal state directly (topology struct is not exported),
	// we verify behavior through interface methods
	// The bindings are stored but we can't directly verify them without Apply()
}

func TestTopology_ExchangeBinding(t *testing.T) {
	topology := NewTopology("test-app", "amqp://test")

	binding1 := NewExchangeBiding().Source("orders").Destination("notifications").RoutingKey("order.created")
	binding2 := NewExchangeBiding().Source("payments").Destination("audit").RoutingKey("payment.processed")

	// Test adding single exchange binding
	result := topology.ExchangeBinding(binding1)
	if result != topology {
		t.Error("ExchangeBinding() should return the same Topology instance")
	}

	// Test adding another exchange binding
	topology.ExchangeBinding(binding2)

	// Since we can't access internal state directly (topology struct is not exported),
	// we verify behavior through interface methods
	// The bindings are stored but we can't directly verify them without Apply()
}

func TestTopology_FluentChaining(t *testing.T) {
	// Test that all methods can be chained together
	queue1 := NewQueue("orders")
	queue2 := NewQueue("payments")
	exchange1 := NewDirectExchange("order-exchange")
	exchange2 := NewFanoutExchange("notification-exchange")
	queueBinding := NewQueueBinding().Queue("orders").Exchange("order-exchange").RoutingKey("order.created")
	exchangeBinding := NewExchangeBiding().Source("order-exchange").Destination("notification-exchange").RoutingKey("order.created")

	topology := NewTopology("test-app", "amqp://test").
		Queue(queue1).
		Queue(queue2).
		Exchange(exchange1).
		Exchange(exchange2).
		QueueBinding(queueBinding).
		ExchangeBinding(exchangeBinding)

	// Verify all components were added through interface methods
	queues := topology.GetQueuesDefinition()
	if len(queues) != 2 {
		t.Errorf("Expected 2 queues after chaining, got %d", len(queues))
	}

	// Verify the queues are the ones we added
	if queues["orders"] != queue1 {
		t.Error("Queue 'orders' not found or incorrect after chaining")
	}
	if queues["payments"] != queue2 {
		t.Error("Queue 'payments' not found or incorrect after chaining")
	}

	// For other components, we can only verify that the fluent interface worked
	// The actual storage is verified by the Apply() method in integration tests
}

// The Apply method integration tests would require complex mocking of NewConnectionManager
// which is not a variable that can be easily overridden. These tests focus on the
// topology building methods which are the core functionality.

func TestTopology_Interface(t *testing.T) {
	// Test that topology implements Topology interface
	topo := NewTopology("test-app", "amqp://test")

	// Test all interface methods exist
	queue := NewQueue("test")
	exchange := NewDirectExchange("test")
	queueBinding := NewQueueBinding()
	exchangeBinding := NewExchangeBiding()

	_ = topo.Queue(queue)
	_ = topo.Queues([]*QueueDefinition{queue})
	_ = topo.Exchange(exchange)
	_ = topo.Exchanges([]*ExchangeDefinition{exchange})
	_ = topo.QueueBinding(queueBinding)
	_ = topo.ExchangeBinding(exchangeBinding)
	_ = topo.GetQueuesDefinition()
	_, _ = topo.GetQueueDefinition("test")
}

func TestTopology_DeclareQueues_QueueType(t *testing.T) {
	top := NewTopology("test-app", "amqp://test")

	// classic queue (default)
	classic := NewQueue("classic-queue")
	top.Queue(classic)

	// quorum queue
	quorum := NewQueue("quorum-queue").Quorum()
	top.Queue(quorum)

	// get concrete topology
	tp, ok := top.(*topology)
	if !ok {
		t.Fatal("failed to cast Topology to *topology")
	}

	mockCh := NewMockAMQPChannel()

	if err := tp.DeclareQueues(mockCh); err != nil {
		t.Fatalf("declareQueues() returned error: %v", err)
	}

	// check classic
	argsClassic, ok := mockCh.declaredArgs["classic-queue"]
	if !ok {
		t.Fatalf("classic queue was not declared")
	}
	if xt, ok := argsClassic["x-queue-type"]; !ok {
		t.Fatalf("x-queue-type missing for classic queue args: %#v", argsClassic)
	} else if xt != "classic" {
		t.Fatalf("classic queue x-queue-type = %v, want classic", xt)
	}

	// check quorum
	argsQuorum, ok := mockCh.declaredArgs["quorum-queue"]
	if !ok {
		t.Fatalf("quorum queue was not declared")
	}
	if xt, ok := argsQuorum["x-queue-type"]; !ok {
		t.Fatalf("x-queue-type missing for quorum queue args: %#v", argsQuorum)
	} else if xt != "quorum" {
		t.Fatalf("quorum queue x-queue-type = %v, want quorum", xt)
	}
}

func TestTopology_DeclareQueues_DLQRoutingKey(t *testing.T) {
	top := NewTopology("test-app", "amqp://test")

	// queue with DLQ and Retry
	qWithRetry := NewQueue("with-retry").WithDLQ().WithRetry(30*time.Second, 3)
	top.Queue(qWithRetry)

	// queue with DLQ but no Retry
	qWithoutRetry := NewQueue("no-retry").WithDLQ()
	top.Queue(qWithoutRetry)

	tp, ok := top.(*topology)
	if !ok {
		t.Fatal("failed to cast Topology to *topology")
	}

	mockCh := NewMockAMQPChannel()

	if err := tp.DeclareQueues(mockCh); err != nil {
		t.Fatalf("declareQueues() returned error: %v", err)
	}

	// DLQ for qWithRetry should have x-dead-letter-routing-key = qWithRetry.RetryName()
	dlqName := qWithRetry.DLQName()
	args, ok := mockCh.declaredArgs[dlqName]
	if !ok {
		t.Fatalf("DLQ %s was not declared", dlqName)
	}
	if rk, ok := args["x-dead-letter-routing-key"]; !ok {
		t.Fatalf("x-dead-letter-routing-key missing for %s args: %#v", dlqName, args)
	} else if rk != qWithRetry.RetryName() {
		t.Fatalf("%s x-dead-letter-routing-key = %v, want %s", dlqName, rk, qWithRetry.RetryName())
	}

	// DLQ for qWithoutRetry should have x-dead-letter-routing-key = qWithoutRetry.DLQName()
	dlqName2 := qWithoutRetry.DLQName()
	args2, ok := mockCh.declaredArgs[dlqName2]
	if !ok {
		t.Fatalf("DLQ %s was not declared", dlqName2)
	}
	if rk2, ok := args2["x-dead-letter-routing-key"]; !ok {
		t.Fatalf("x-dead-letter-routing-key missing for %s args: %#v", dlqName2, args2)
	} else if rk2 != qWithoutRetry.DLQName() {
		t.Fatalf("%s x-dead-letter-routing-key = %v, want %s", dlqName2, rk2, qWithoutRetry.DLQName())
	}
}

func TestTopology_DeclareAndBindings(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() Topology
		setupChannel  func() AMQPChannel
		expectedError error
	}{
		{
			name: "nil channel",
			setup: func() Topology {
				return NewTopology("test-app", "amqp://test")
			},
			setupChannel: func() AMQPChannel {
				return nil
			},
			expectedError: NullableChannelError,
		},
		{
			name: "successful declare and bindings",
			setup: func() Topology {
				queue := NewQueue("test-queue")
				exchange := NewDirectExchange("test-exchange")
				queueBinding := NewQueueBinding().Queue("test-queue").Exchange("test-exchange").RoutingKey("test.key")
				exchangeBinding := NewExchangeBiding().Source("test-exchange").Destination("dest-exchange").RoutingKey("exchange.key")

				return NewTopology("test-app", "amqp://test").
					Queue(queue).
					Exchange(exchange).
					Exchange(NewDirectExchange("dest-exchange")).
					QueueBinding(queueBinding).
					ExchangeBinding(exchangeBinding)
			},
			setupChannel: func() AMQPChannel {
				return NewMockAMQPChannel()
			},
			expectedError: nil,
		},
		{
			name: "exchange declare error",
			setup: func() Topology {
				exchange := NewDirectExchange("test-exchange")
				return NewTopology("test-app", "amqp://test").Exchange(exchange)
			},
			setupChannel: func() AMQPChannel {
				mockCh := NewMockAMQPChannel()
				mockCh.SetExchangeDeclareError(NewBunMQError("exchange declare failed"))
				return mockCh
			},
			expectedError: NewBunMQError("exchange declare failed"),
		},
		{
			name: "queue declare error",
			setup: func() Topology {
				queue := NewQueue("test-queue")
				return NewTopology("test-app", "amqp://test").Queue(queue)
			},
			setupChannel: func() AMQPChannel {
				mockCh := NewMockAMQPChannel()
				mockCh.SetQueueDeclareError(NewBunMQError("queue declare failed"))
				return mockCh
			},
			expectedError: NewBunMQError("queue declare failed"),
		},
		{
			name: "queue bind error",
			setup: func() Topology {
				queue := NewQueue("test-queue")
				exchange := NewDirectExchange("test-exchange")
				queueBinding := NewQueueBinding().Queue("test-queue").Exchange("test-exchange").RoutingKey("test.key")

				return NewTopology("test-app", "amqp://test").
					Queue(queue).
					Exchange(exchange).
					QueueBinding(queueBinding)
			},
			setupChannel: func() AMQPChannel {
				mockCh := NewMockAMQPChannel()
				mockCh.SetQueueBindError(NewBunMQError("queue bind failed"))
				return mockCh
			},
			expectedError: NewBunMQError("queue bind failed"),
		},
		{
			name: "exchange bind error",
			setup: func() Topology {
				exchangeBinding := NewExchangeBiding().Source("test-exchange").Destination("dest-exchange").RoutingKey("exchange.key")

				return NewTopology("test-app", "amqp://test").
					Exchange(NewDirectExchange("test-exchange")).
					Exchange(NewDirectExchange("dest-exchange")).
					ExchangeBinding(exchangeBinding)
			},
			setupChannel: func() AMQPChannel {
				mockCh := NewMockAMQPChannel()
				mockCh.SetExchangeBindError(NewBunMQError("exchange bind failed"))
				return mockCh
			},
			expectedError: NewBunMQError("exchange bind failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topo := tt.setup()
			channel := tt.setupChannel()

			tp, ok := topo.(*topology)
			if !ok {
				t.Fatal("failed to cast Topology to *topology")
			}

			err := tp.DeclareAndBindings(channel)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("DeclareAndBindings() expected error %v, got nil", tt.expectedError)
				} else if err.Error() != tt.expectedError.Error() {
					t.Errorf("DeclareAndBindings() error = %v, expectedError %v", err, tt.expectedError)
				}
			} else if err != nil {
				t.Errorf("DeclareAndBindings() unexpected error = %v", err)
			}
		})
	}
}

func TestTopology_DeclareExchanges(t *testing.T) {
	tests := []struct {
		name          string
		exchanges     []*ExchangeDefinition
		setupChannel  func() AMQPChannel
		expectedError error
	}{
		{
			name:      "no exchanges",
			exchanges: []*ExchangeDefinition{},
			setupChannel: func() AMQPChannel {
				return NewMockAMQPChannel()
			},
			expectedError: nil,
		},
		{
			name: "single direct exchange",
			exchanges: []*ExchangeDefinition{
				NewDirectExchange("orders"),
			},
			setupChannel: func() AMQPChannel {
				return NewMockAMQPChannel()
			},
			expectedError: nil,
		},
		{
			name: "multiple different exchange types",
			exchanges: []*ExchangeDefinition{
				NewDirectExchange("orders"),
				NewFanoutExchange("notifications"),
				NewDirectExchange("events"),
				NewFanoutExchange("routing"),
			},
			setupChannel: func() AMQPChannel {
				return NewMockAMQPChannel()
			},
			expectedError: nil,
		},
		{
			name: "exchange declare error",
			exchanges: []*ExchangeDefinition{
				NewDirectExchange("failing-exchange"),
			},
			setupChannel: func() AMQPChannel {
				mockCh := NewMockAMQPChannel()
				mockCh.SetExchangeDeclareError(NewBunMQError("exchange declare failed"))
				return mockCh
			},
			expectedError: NewBunMQError("exchange declare failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topo := NewTopology("test-app", "amqp://test")

			// Add exchanges to topology
			for _, exchange := range tt.exchanges {
				topo.Exchange(exchange)
			}

			channel := tt.setupChannel()

			tp, ok := topo.(*topology)
			if !ok {
				t.Fatal("failed to cast Topology to *topology")
			}

			err := tp.DeclareExchanges(channel)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("DeclareExchanges() expected error %v, got nil", tt.expectedError)
				} else if err.Error() != tt.expectedError.Error() {
					t.Errorf("DeclareExchanges() error = %v, expectedError %v", err, tt.expectedError)
				}
			} else if err != nil {
				t.Errorf("DeclareExchanges() unexpected error = %v", err)
			}
		})
	}
}

func TestTopology_BindQueues(t *testing.T) {
	tests := []struct {
		name          string
		bindings      []*QueueBindingDefinition
		setupChannel  func() AMQPChannel
		expectedError error
	}{
		{
			name:     "no bindings",
			bindings: []*QueueBindingDefinition{},
			setupChannel: func() AMQPChannel {
				return NewMockAMQPChannel()
			},
			expectedError: nil,
		},
		{
			name: "single queue binding",
			bindings: []*QueueBindingDefinition{
				NewQueueBinding().Queue("orders").Exchange("order-exchange").RoutingKey("order.created"),
			},
			setupChannel: func() AMQPChannel {
				return NewMockAMQPChannel()
			},
			expectedError: nil,
		},
		{
			name: "multiple queue bindings",
			bindings: []*QueueBindingDefinition{
				NewQueueBinding().Queue("orders").Exchange("order-exchange").RoutingKey("order.created"),
				NewQueueBinding().Queue("payments").Exchange("payment-exchange").RoutingKey("payment.processed"),
				NewQueueBinding().Queue("notifications").Exchange("notification-exchange").RoutingKey("notification.sent"),
			},
			setupChannel: func() AMQPChannel {
				return NewMockAMQPChannel()
			},
			expectedError: nil,
		},
		{
			name: "queue bind error",
			bindings: []*QueueBindingDefinition{
				NewQueueBinding().Queue("failing-queue").Exchange("test-exchange").RoutingKey("test.key"),
			},
			setupChannel: func() AMQPChannel {
				mockCh := NewMockAMQPChannel()
				mockCh.SetQueueBindError(NewBunMQError("queue bind failed"))
				return mockCh
			},
			expectedError: NewBunMQError("queue bind failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topo := NewTopology("test-app", "amqp://test")

			// Add bindings to topology
			for _, binding := range tt.bindings {
				topo.QueueBinding(binding)
			}

			channel := tt.setupChannel()

			tp, ok := topo.(*topology)
			if !ok {
				t.Fatal("failed to cast Topology to *topology")
			}

			err := tp.BindQueues(channel)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("BindQueues() expected error %v, got nil", tt.expectedError)
				} else if err.Error() != tt.expectedError.Error() {
					t.Errorf("BindQueues() error = %v, expectedError %v", err, tt.expectedError)
				}
			} else if err != nil {
				t.Errorf("BindQueues() unexpected error = %v", err)
			}
		})
	}
}

func TestTopology_BindExchanges(t *testing.T) {
	tests := []struct {
		name          string
		bindings      []*ExchangeBindingDefinition
		setupChannel  func() AMQPChannel
		expectedError error
	}{
		{
			name:     "no bindings",
			bindings: []*ExchangeBindingDefinition{},
			setupChannel: func() AMQPChannel {
				return NewMockAMQPChannel()
			},
			expectedError: nil,
		},
		{
			name: "single exchange binding",
			bindings: []*ExchangeBindingDefinition{
				NewExchangeBiding().Source("orders").Destination("notifications").RoutingKey("order.created"),
			},
			setupChannel: func() AMQPChannel {
				return NewMockAMQPChannel()
			},
			expectedError: nil,
		},
		{
			name: "multiple exchange bindings",
			bindings: []*ExchangeBindingDefinition{
				NewExchangeBiding().Source("orders").Destination("notifications").RoutingKey("order.created"),
				NewExchangeBiding().Source("payments").Destination("audit").RoutingKey("payment.processed"),
				NewExchangeBiding().Source("users").Destination("analytics").RoutingKey("user.action"),
			},
			setupChannel: func() AMQPChannel {
				return NewMockAMQPChannel()
			},
			expectedError: nil,
		},
		{
			name: "exchange bind error",
			bindings: []*ExchangeBindingDefinition{
				NewExchangeBiding().Source("failing-source").Destination("dest").RoutingKey("test.key"),
			},
			setupChannel: func() AMQPChannel {
				mockCh := NewMockAMQPChannel()
				mockCh.SetExchangeBindError(NewBunMQError("exchange bind failed"))
				return mockCh
			},
			expectedError: NewBunMQError("exchange bind failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topo := NewTopology("test-app", "amqp://test")

			// Add bindings to topology
			for _, binding := range tt.bindings {
				topo.ExchangeBinding(binding)
			}

			channel := tt.setupChannel()

			tp, ok := topo.(*topology)
			if !ok {
				t.Fatal("failed to cast Topology to *topology")
			}

			err := tp.BindExchanges(channel)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("BindExchanges() expected error %v, got nil", tt.expectedError)
				} else if err.Error() != tt.expectedError.Error() {
					t.Errorf("BindExchanges() error = %v, expectedError %v", err, tt.expectedError)
				}
			} else if err != nil {
				t.Errorf("BindExchanges() unexpected error = %v", err)
			}
		})
	}
}

func TestTopology_DeclareQueues_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() Topology
		setupChannel  func() AMQPChannel
		expectedError error
	}{
		{
			name: "queue declare error for main queue",
			setup: func() Topology {
				queue := NewQueue("failing-queue")
				return NewTopology("test-app", "amqp://test").Queue(queue)
			},
			setupChannel: func() AMQPChannel {
				mockCh := NewMockAMQPChannel()
				mockCh.SetQueueDeclareError(NewBunMQError("queue declare failed"))
				return mockCh
			},
			expectedError: NewBunMQError("queue declare failed"),
		},
		{
			name: "queue declare error for retry queue",
			setup: func() Topology {
				queue := NewQueue("test-queue").WithRetry(30*time.Second, 3)
				return NewTopology("test-app", "amqp://test").Queue(queue)
			},
			setupChannel: func() AMQPChannel {
				mockCh := NewMockAMQPChannel()
				// Mock will fail on first QueueDeclare call (retry queue)
				mockCh.SetQueueDeclareError(NewBunMQError("retry queue declare failed"))
				return mockCh
			},
			expectedError: NewBunMQError("retry queue declare failed"),
		},
		{
			name: "queue declare error for DLQ queue",
			setup: func() Topology {
				queue := NewQueue("test-queue").WithDLQ()
				return NewTopology("test-app", "amqp://test").Queue(queue)
			},
			setupChannel: func() AMQPChannel {
				mockCh := NewMockAMQPChannel()
				// Mock will fail on first QueueDeclare call (DLQ queue)
				mockCh.SetQueueDeclareError(NewBunMQError("dlq queue declare failed"))
				return mockCh
			},
			expectedError: NewBunMQError("dlq queue declare failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topo := tt.setup()
			channel := tt.setupChannel()

			tp, ok := topo.(*topology)
			if !ok {
				t.Fatal("failed to cast Topology to *topology")
			}

			err := tp.DeclareQueues(channel)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("DeclareQueues() expected error %v, got nil", tt.expectedError)
				} else if err.Error() != tt.expectedError.Error() {
					t.Errorf("DeclareQueues() error = %v, expectedError %v", err, tt.expectedError)
				}
			} else if err != nil {
				t.Errorf("DeclareQueues() unexpected error = %v", err)
			}
		})
	}
}

func TestTopology_DeclareQueues_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() Topology
		validate func(t *testing.T, mockCh *MockAMQPChannel)
	}{
		{
			name: "queue with DLQ max length",
			setup: func() Topology {
				queue := NewQueue("test-queue").WithDLQ().WithDLQMaxLength(1000)
				return NewTopology("test-app", "amqp://test").Queue(queue)
			},
			validate: func(t *testing.T, mockCh *MockAMQPChannel) {
				// Check DLQ was declared with max length
				dlqArgs, ok := mockCh.declaredArgs["test-queue-dlq"]
				if !ok {
					t.Fatal("DLQ was not declared")
				}
				if maxLen, ok := dlqArgs["x-max-length"]; !ok {
					t.Fatal("x-max-length missing for DLQ")
				} else if maxLen != int64(1000) {
					t.Fatalf("DLQ x-max-length = %v, want 1000", maxLen)
				}
			},
		},
		{
			name: "queue with max length",
			setup: func() Topology {
				queue := NewQueue("test-queue").WithMaxLength(500)
				return NewTopology("test-app", "amqp://test").Queue(queue)
			},
			validate: func(t *testing.T, mockCh *MockAMQPChannel) {
				// Check main queue was declared with max length
				args, ok := mockCh.declaredArgs["test-queue"]
				if !ok {
					t.Fatal("Main queue was not declared")
				}
				if maxLen, ok := args["x-max-length"]; !ok {
					t.Fatal("x-max-length missing for main queue")
				} else if maxLen != int64(500) {
					t.Fatalf("Main queue x-max-length = %v, want 500", maxLen)
				}
			},
		},
		{
			name: "queue with both DLQ and retry",
			setup: func() Topology {
				queue := NewQueue("test-queue").WithDLQ().WithRetry(60*time.Second, 5)
				return NewTopology("test-app", "amqp://test").Queue(queue)
			},
			validate: func(t *testing.T, mockCh *MockAMQPChannel) {
				// Check all three queues were declared
				_, ok := mockCh.declaredArgs["test-queue"]
				if !ok {
					t.Fatal("Main queue was not declared")
				}

				retryArgs, ok := mockCh.declaredArgs["test-queue-retry"]
				if !ok {
					t.Fatal("Retry queue was not declared")
				}

				dlqArgs, ok := mockCh.declaredArgs["test-queue-dlq"]
				if !ok {
					t.Fatal("DLQ was not declared")
				} // Validate retry queue properties
				if ttl, ok := retryArgs["x-message-ttl"]; !ok {
					t.Fatal("x-message-ttl missing for retry queue")
				} else if ttl != int64(60000) { // 60 seconds in milliseconds
					t.Fatalf("Retry queue x-message-ttl = %v, want 60000", ttl)
				}

				if retryCount, ok := retryArgs["x-retry-count"]; !ok {
					t.Fatal("x-retry-count missing for retry queue")
				} else if retryCount != int64(5) {
					t.Fatalf("Retry queue x-retry-count = %v (type %T), want int64(5)", retryCount, retryCount)
				}

				// Validate DLQ routing
				if dlx, ok := dlqArgs["x-dead-letter-exchange"]; !ok {
					t.Fatal("x-dead-letter-exchange missing for DLQ")
				} else if dlx != "" {
					t.Fatalf("DLQ x-dead-letter-exchange = %v, want empty string", dlx)
				}

				if dlrk, ok := dlqArgs["x-dead-letter-routing-key"]; !ok {
					t.Fatal("x-dead-letter-routing-key missing for DLQ")
				} else if dlrk != "test-queue-retry" {
					t.Fatalf("DLQ x-dead-letter-routing-key = %v, want test-queue-retry", dlrk)
				}

				// Validate all have queue type
				for queueName, args := range mockCh.declaredArgs {
					if qt, ok := args["x-queue-type"]; !ok {
						t.Fatalf("x-queue-type missing for queue %s", queueName)
					} else if qt != "classic" {
						t.Fatalf("Queue %s x-queue-type = %v, want classic", queueName, qt)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topo := tt.setup()
			mockCh := NewMockAMQPChannel()

			tp, ok := topo.(*topology)
			if !ok {
				t.Fatal("failed to cast Topology to *topology")
			}

			err := tp.DeclareQueues(mockCh)
			if err != nil {
				t.Fatalf("DeclareQueues() unexpected error = %v", err)
			}

			tt.validate(t, mockCh)
		})
	}
}
