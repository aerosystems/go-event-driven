package main

import (
	"time"
)

const defaultRetriesCount = 10

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

	newsletterErr := make(chan error)
	go exponentialBackoffAsync(h.newsletterClient.AddToNewsletter, u, newsletterErr)

	notificationsErr := make(chan error)
	go exponentialBackoffAsync(h.notificationsClient.SendNotification, u, notificationsErr)

	if err := <-newsletterErr; err != nil {
		return err
	}

	if err := <-notificationsErr; err != nil {
		return err
	}
	return nil
}

func exponentialBackoffAsync(f func(u User) error, u User, errChan chan<- error) {
	errChan <- exponentialBackoff(f, u)
}

func exponentialBackoff(f func(u User) error, u User) error {
	var err error
	delay := 1
	term := 1
	for i := 1; i <= defaultRetriesCount; i++ {
		err = f(u)
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(delay) * time.Second)
		term *= defaultRetriesCount / i
		delay += term
	}
	return err
}
