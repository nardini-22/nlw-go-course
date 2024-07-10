package mailpit

import (
	"context"
	"fmt"
	"nlw-go-course/internal/pgstore"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wneessen/go-mail"
)

type store interface {
	GetTrip(context.Context, uuid.UUID) (pgstore.Trip, error)
}

type Mailpit struct {
	store store
}

func NewMailpit(pool *pgxpool.Pool) Mailpit {
	return Mailpit{pgstore.New(pool)}
}

func (mp Mailpit) SendConfirmTripEmailToTripOwner(tripID uuid.UUID) error {
	ctx := context.Background()
	trip, err := mp.store.GetTrip(ctx, tripID)
	if err != nil {
		return fmt.Errorf("Mailpit failed to get trip for SendConfirmTripEmailToTripOwner: %w", err)
	}

	msg := mail.NewMsg()
	if err := msg.From("mailpit@course.com"); err != nil {
		return fmt.Errorf("Mailpit failed to set From in email: %w", err)
	}

	if err := msg.To(trip.OwnerEmail); err != nil {
		return fmt.Errorf("Mailpit failed to set To in email: %w", err)
	}

	msg.Subject("Confirm your trip")
	msg.SetBodyString(mail.TypeTextPlain, fmt.Sprintf(
		`Hello, %s!

		Your trip to %s which begins on %s needs to be confirmed.
		Click on the button below to confirm.
		`, 
		trip.OwnerName, trip.Destination, trip.StartsAt.Time.Format(time.DateOnly),
	))

	client, err := mail.NewClient("mailpit", mail.WithTLSPortPolicy(mail.NoTLS), mail.WithPort(1025))
	if err != nil {
		return fmt.Errorf("Mailpit failed to create email client: %w", err)
	}

	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("Mailpit failed to send email: %w", err)
	}

	return nil

}