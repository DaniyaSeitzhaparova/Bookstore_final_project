package handler

import (
	"context"
	"encoding/json"
	"log"

	"github.com/OshakbayAigerim/read_space/notification_service/internal/domain" // <- сюда
	"github.com/OshakbayAigerim/read_space/notification_service/internal/usecase"
	"github.com/nats-io/nats.go"
)

func SubscribeAll(nc *nats.Conn, notifier *usecase.Notifier) error {
	if _, err := nc.Subscribe("orders.created", func(m *nats.Msg) {
		var evt domain.OrderCreatedEvent
		if err := json.Unmarshal(m.Data, &evt); err != nil {
			log.Printf(" unmarshal orders.created: %v", err)
			return
		}
		notifier.SendOrderConfirmation(context.Background(), evt)
	}); err != nil {
		return err
	}

	if _, err := nc.Subscribe("user.created", func(m *nats.Msg) {
		var evt domain.UserCreatedEvent
		if err := json.Unmarshal(m.Data, &evt); err != nil {
			log.Printf("unmarshal user.created: %v", err)
			return
		}
		notifier.SendWelcome(context.Background(), evt)
	}); err != nil {
		return err
	}

	if _, err := nc.Subscribe("order.completed", func(m *nats.Msg) {
		var evt domain.OrderCompletedEvent
		if err := json.Unmarshal(m.Data, &evt); err != nil {
			log.Printf("unmarshal order.completed: %v", err)
			return
		}
		notifier.SendOrderCompleted(context.Background(), evt)
	}); err != nil {
		return err
	}

	if _, err := nc.Subscribe("order.deleted", func(m *nats.Msg) {
		var evt domain.OrderDeletedEvent
		if err := json.Unmarshal(m.Data, &evt); err != nil {
			log.Printf("unmarshal order.deleted: %v", err)
			return
		}
		notifier.SendOrderDeleted(context.Background(), evt)
	}); err != nil {
		return err
	}

	if _, err := nc.Subscribe("exchange.offered", func(m *nats.Msg) {
		var evt domain.OfferCreatedEvent
		if err := json.Unmarshal(m.Data, &evt); err != nil {
			log.Printf(" unmarshal exchange.offered: %v", err)
			return
		}
		notifier.SendOfferCreated(context.Background(), evt)
	}); err != nil {
		return err
	}

	if _, err := nc.Subscribe("exchange.accepted", func(m *nats.Msg) {
		var evt domain.OfferAcceptedEvent
		if err := json.Unmarshal(m.Data, &evt); err != nil {
			log.Printf(" unmarshal exchange.accepted: %v", err)
			return
		}
		notifier.SendOfferAccepted(context.Background(), evt)
	}); err != nil {
		return err
	}

	if _, err := nc.Subscribe("exchange.declined", func(m *nats.Msg) {
		var evt domain.OfferDeclinedEvent
		if err := json.Unmarshal(m.Data, &evt); err != nil {
			log.Printf("unmarshal exchange.declined: %v", err)
			return
		}
		notifier.SendOfferDeclined(context.Background(), evt)
	}); err != nil {
		return err
	}

	if _, err := nc.Subscribe("userlibrary.book.assigned", func(m *nats.Msg) {
		var evt domain.BookAssignedEvent
		if err := json.Unmarshal(m.Data, &evt); err == nil {
			notifier.SendBookAssigned(context.Background(), evt)
		}
	}); err != nil {
		return err
	}

	if _, err := nc.Subscribe("userlibrary.book.unassigned", func(m *nats.Msg) {
		var evt domain.BookUnassignedEvent
		if err := json.Unmarshal(m.Data, &evt); err == nil {
			notifier.SendBookUnassigned(context.Background(), evt)
		}
	}); err != nil {
		return err
	}

	if _, err := nc.Subscribe("userlibrary.entry.deleted", func(m *nats.Msg) {
		var evt domain.EntryDeletedEvent
		if err := json.Unmarshal(m.Data, &evt); err == nil {
			notifier.SendEntryDeleted(context.Background(), evt)
		}
	}); err != nil {
		return err
	}

	if _, err := nc.Subscribe("userlibrary.entry.updated", func(m *nats.Msg) {
		var evt domain.EntryUpdatedEvent
		if err := json.Unmarshal(m.Data, &evt); err == nil {
			notifier.SendEntryUpdated(context.Background(), evt)
		}
	}); err != nil {
		return err
	}

	return nil
}
