package main

import "log/slog"

type ConsoleConsumer struct {
	*slog.Logger
}

func (c *ConsoleConsumer) BatchAccept(list []*EventData) {
	for _, data := range list {
		c.Info("Console Received :",
			slog.String("Action", data.Action),
			slog.Any("PKColumns", data.PKColumns),
			slog.Any("Before", data.Before),
			slog.Any("After", data.After),
		)
	}
}

func (c *ConsoleConsumer) Accept(data *EventData) {
	c.BatchAccept([]*EventData{data})
}
