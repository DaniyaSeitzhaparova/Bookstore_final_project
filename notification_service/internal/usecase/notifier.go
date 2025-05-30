package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/OshakbayAigerim/read_space/notification_service/internal/domain"
	userpb "github.com/OshakbayAigerim/read_space/user_service/proto"
)

type EmailSender func(to, subject, body string)

type Notifier struct {
	userClient userpb.UserServiceClient
	sendEmail  EmailSender
}

func NewNotifier(userClient userpb.UserServiceClient, sendEmail EmailSender) *Notifier {
	return &Notifier{userClient: userClient, sendEmail: sendEmail}
}

func (n *Notifier) getEmail(ctx context.Context, userID string) (string, error) {
	resp, err := n.userClient.GetUser(ctx, &userpb.UserID{Id: userID})
	if err != nil {
		return "", fmt.Errorf("grpc GetUser: %w", err)
	}
	return resp.User.Email, nil
}

func (n *Notifier) SendOrderConfirmation(ctx context.Context, evt domain.OrderCreatedEvent) {
	email, err := n.getEmail(ctx, evt.UserID)
	if err != nil {
		log.Printf(" cannot fetch email for %s: %v", evt.UserID, err)
		return
	}
	subject := "Ваш заказ оформлен"
	body := fmt.Sprintf("Спасибо за заказ %s!", evt.OrderID)
	n.sendEmail(email, subject, body)
	log.Printf(" Email sent to %s", email)
}

func (n *Notifier) SendWelcome(ctx context.Context, evt domain.UserCreatedEvent) {
	subject := "Добро пожаловать в ReadSpace!"
	body := fmt.Sprintf("Привет, %s!\n\nСпасибо за регистрацию.", evt.Name)
	n.sendEmail(evt.Email, subject, body)
	log.Printf(" Welcome email sent to %s", evt.Email)
}

func (n *Notifier) SendOrderCompleted(ctx context.Context, evt domain.OrderCompletedEvent) {
	email, err := n.getEmail(ctx, evt.UserID)
	if err != nil {
		log.Printf(" cannot fetch email for %s: %v", evt.UserID, err)
		return
	}
	subject := "Ваш заказ возвращён"
	body := fmt.Sprintf("Заказ %s помечен как возвращён.", evt.OrderID)
	n.sendEmail(email, subject, body)
	log.Printf("Email sent to %s", email)
}

func (n *Notifier) SendOrderDeleted(ctx context.Context, evt domain.OrderDeletedEvent) {
	email, err := n.getEmail(ctx, evt.UserID)
	if err != nil {
		log.Printf(" cannot fetch email for %s: %v", evt.UserID, err)
		return
	}
	subject := "Ваш заказ удалён"
	body := fmt.Sprintf("Заказ %s был удалён.", evt.OrderID)
	n.sendEmail(email, subject, body)
	log.Printf(" Email sent to %s", email)
}

func (n *Notifier) SendOfferCreated(ctx context.Context, evt domain.OfferCreatedEvent) {
	email, err := n.getEmail(ctx, evt.OwnerID)
	if err != nil {
		log.Printf(" cannot fetch email for %s: %v", evt.OwnerID, err)
		return
	}
	subject := "Поступило новое предложение обмена"
	body := fmt.Sprintf("Ваше предложение %s создано.", evt.OfferID)
	n.sendEmail(email, subject, body)
	log.Printf(" Email sent to %s", email)
}

func (n *Notifier) SendOfferDeclined(ctx context.Context, evt domain.OfferDeclinedEvent) {
	email, err := n.getEmail(ctx, evt.OwnerID)
	if err != nil {
		log.Printf(" cannot fetch email for %s: %v", evt.OwnerID, err)
		return
	}
	subject := "Ваше предложение обмена отклонено"
	body := fmt.Sprintf("Предложение %s было отклонено.", evt.OfferID)
	n.sendEmail(email, subject, body)
	log.Printf(" Email sent to %s", email)
}
func (n *Notifier) SendOfferAccepted(ctx context.Context, evt domain.OfferAcceptedEvent) {
	email, err := n.getEmail(ctx, evt.OwnerID)
	if err != nil {
		log.Printf("cannot fetch email for %s: %v", evt.OwnerID, err)
		return
	}
	subject := "Ваше предложение обмена принято"
	body := fmt.Sprintf("Предложение %s принято пользователем %s.", evt.OfferID, evt.Counterparty)
	n.sendEmail(email, subject, body)
	log.Printf(" Email sent to %s", email)
}

func (n *Notifier) SendBookAssigned(ctx context.Context, evt domain.BookAssignedEvent) {
	email, err := n.getEmail(ctx, evt.UserID)
	if err != nil {
		log.Printf("cannot fetch email for %s: %v", evt.UserID, err)
		return
	}
	subject := "Book Assigned"
	body := fmt.Sprintf("The book %s has been assigned to you.", evt.BookID)
	n.sendEmail(email, subject, body)
}

func (n *Notifier) SendBookUnassigned(ctx context.Context, evt domain.BookUnassignedEvent) {
	email, err := n.getEmail(ctx, evt.UserID)
	if err != nil {
		log.Printf("cannot fetch email for %s: %v", evt.UserID, err)
		return
	}
	subject := "Book Unassigned"
	body := fmt.Sprintf("The book %s has been unassigned from you.", evt.BookID)
	n.sendEmail(email, subject, body)
}

func (n *Notifier) SendEntryDeleted(ctx context.Context, evt domain.EntryDeletedEvent) {
	email, err := n.getEmail(ctx, evt.UserID)
	if err != nil {
		log.Printf("cannot fetch email for %s: %v", evt.UserID, err)
		return
	}
	subject := "Library Entry Deleted"
	body := fmt.Sprintf("Your library entry %s was deleted.", evt.EntryID)
	n.sendEmail(email, subject, body)
}

func (n *Notifier) SendEntryUpdated(ctx context.Context, evt domain.EntryUpdatedEvent) {
	email, err := n.getEmail(ctx, evt.UserID)
	if err != nil {
		log.Printf("cannot fetch email for %s: %v", evt.UserID, err)
		return
	}
	subject := "Library Entry Updated"
	body := fmt.Sprintf("Your library entry %s was updated (new book %s).", evt.EntryID, evt.BookID)
	n.sendEmail(email, subject, body)
}
