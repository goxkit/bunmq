// Copyright (c) 2025, The GoKit Authors
// MIT License
// All rights reserved.

package bunmq

type (
	OptionKey string

	Option struct {
		Key   OptionKey
		Value any
	}

	DeliveryMode uint8
)

const (
	OptionDeliveryModeKey OptionKey = "DeliveryMode"
	OptionHeadersKey      OptionKey = "Headers"

	DeliveryModeTransient  DeliveryMode = 1
	DeliveryModePersistent DeliveryMode = 2
)

func OptionPersistentDeliveryMode() []*Option {
	return []*Option{{Key: OptionDeliveryModeKey, Value: DeliveryModePersistent}}
}

func OptionTransientDeliveryMode() []*Option {
	return []*Option{{Key: OptionDeliveryModeKey, Value: DeliveryModeTransient}}
}

func OptionHeaders(headers map[string]any) []*Option {
	return []*Option{{Key: OptionHeadersKey, Value: headers}}
}

func PublisherOptions(deliveryMode DeliveryMode, headers map[string]any) []*Option {
	return []*Option{
		{Key: OptionDeliveryModeKey, Value: deliveryMode},
		{Key: OptionHeadersKey, Value: headers},
	}
}
