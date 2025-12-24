// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

import "github.com/rabbitmq/amqp091-go"

type (
	OptionKey string

	Option struct {
		Key   OptionKey
		Value any
	}

	OptionsBuilder struct {
		options []*Option
	}
)

const (
	OptionDeliveryModeKey OptionKey = "DeliveryMode"
	OptionHeadersKey      OptionKey = "Headers"
)

func NewOption() *OptionsBuilder {
	return &OptionsBuilder{options: []*Option{}}
}

func (b *OptionsBuilder) WithOption(option *Option) *OptionsBuilder {
	b.options = append(b.options, option)
	return b
}

func (b *OptionsBuilder) WithDeliveryMode(deliveryMode uint8) *OptionsBuilder {
	b.options = append(b.options, &Option{Key: OptionDeliveryModeKey, Value: deliveryMode})
	return b
}

func (b *OptionsBuilder) WithDeliveryModeTransient() *OptionsBuilder {
	b.options = append(b.options, &Option{Key: OptionDeliveryModeKey, Value: amqp091.Transient})
	return b
}

func (b *OptionsBuilder) WithDeliveryModePersistent() *OptionsBuilder {
	b.options = append(b.options, &Option{Key: OptionDeliveryModeKey, Value: amqp091.Persistent})
	return b
}

func (b *OptionsBuilder) WithHeaders(headers map[string]any) *OptionsBuilder {
	b.options = append(b.options, &Option{Key: OptionHeadersKey, Value: headers})
	return b
}

func (b *OptionsBuilder) Build() []*Option {
	return b.options
}
