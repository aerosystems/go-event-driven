package main

import (
	"context"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
)

type User struct {
	Email string
}

type UserRepository interface {
	CreateUserAccount(u User) error
}

type NotificationsClient interface {
	SendNotification(u User) error
}

type NewsletterClient interface {
	AddToNewsletter(u User) error
}

type Handler struct {
	repository          UserRepository
	newsletterClient    NewsletterClient
	notificationsClient NotificationsClient
}

func NewHandler(
	repository UserRepository,
	newsletterClient NewsletterClient,
	notificationsClient NotificationsClient,
) Handler {
	return Handler{
		repository:          repository,
		newsletterClient:    newsletterClient,
		notificationsClient: notificationsClient,
	}
}

func (h Handler) SignUp(u User) error {
	if err := h.repository.CreateUserAccount(u); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	go exponentialBackoff(ctx, func() error {
		return h.newsletterClient.AddToNewsletter(u)
	})

	go exponentialBackoff(ctx, func() error {
		return h.notificationsClient.SendNotification(u)
	})

	return nil
}

func exponentialBackoff(ctx context.Context, f func() error) {
	delay := 10 * time.Millisecond
	select {
	case <-ctx.Done():
		return
	default:
		for {
			if err := f(); err == nil {
				return
			}
			time.Sleep(delay)
			delay *= 2
		}
	}
}
