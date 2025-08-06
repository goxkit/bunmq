// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package main

import (
	"context"
	"time"

	"github.com/goxkit/bunmq"
	"github.com/sirupsen/logrus"
)

func main() {
	queueDef := bunmq.
		NewQueue("my-queue").
		Durable(true).
		WithMaxLength(100_000).
		WithRetry(time.Second*10, 3).
		WithDQL().
		WithDLQMaxLength(10_000)

	manager, _ := bunmq.NewConnectionManager("amqp://guest:guest@localhost:5672/")

	topology := bunmq.
		NewTopology(manager).
		Queue(queueDef).
		Exchange(
			bunmq.
				NewDirectExchange("my-exchange").
				Durable(true),
		).
		QueueBinding(
			bunmq.
				NewQueueBinding().
				Queue("my-queue").
				Exchange("my-exchange").
				RoutingKey("my-routing-key"),
		)

	err := topology.Apply()
	if err != nil {
		panic(err)
	}

	dispatcher := bunmq.NewDispatcher(manager, []*bunmq.QueueDefinition{queueDef})

	dispatcher.RegisterByType(
		queueDef.Name(),
		&MyCustomMessage{},
		func(ctx context.Context, msg any, metadata *bunmq.DeliveryMetadata) error {
			logrus.Info("example dispatcher received message:", msg)
			return nil
		})

	dispatcher.ConsumeBlocking()
}

type MyCustomMessage struct {
	Value string `json:"value"`
}
