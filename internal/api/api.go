package api

import (
	"context"
	"errors"
	"net/http"
	"nlw-go-course/internal/api/spec"
	"nlw-go-course/internal/pgstore"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type store interface{
	GetParticipant(ctx context.Context, participantID uuid.UUID) (pgstore.Participant, error)
	ConfirmParticipant(ctx context.Context, participantID uuid.UUID) error
}

type Api struct{
	store store
	logger *zap.Logger
} 

func NewApi(pool *pgxpool.Pool, logger *zap.Logger) Api {
	return Api {pgstore.New(pool), logger}
}

// Confirms a participant on a trip.
// (PATCH /participants/{participantId}/confirm)
func (api *Api) PatchParticipantsParticipantIDConfirm(w http.ResponseWriter, r *http.Request, participantID string) *spec.Response {
	id, err := uuid.Parse(participantID)
	if err != nil {
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "Invalid uuid"})
	}
	participant, err := api.store.GetParticipant(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows){
			return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "Participant not found"})
		}
		api.logger.Error("Failed to get participant", zap.Error(err), zap.String("participant_id",participantID))
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "Something went wrong, try again later"})
	}
	if participant.IsConfirmed {
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "Participant is already confirmed"})
	}
	 if err := api.store.ConfirmParticipant(r.Context(), id); 
	 	err != nil{
		api.logger.Error("Failed to confirm participant", zap.Error(err), zap.String("participant_id",participantID))
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "Something went wrong, try again later"})
	 }
	 return spec.PatchParticipantsParticipantIDConfirmJSON204Response(nil)
}

// Create a new trip
// (POST /trips)
func (api *Api) PostTrips(w http.ResponseWriter, r *http.Request) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Get a trip details.
// (GET /trips/{tripId})
func (api *Api) GetTripsTripID(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Update a trip.
// (PUT /trips/{tripId})
func (api *Api) PutTripsTripID(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Get a trip activities.
// (GET /trips/{tripId}/activities)
func (api *Api) GetTripsTripIDActivities(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Create a trip activity.
// (POST /trips/{tripId}/activities)
func (api *Api) PostTripsTripIDActivities(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Confirm a trip and send e-mail invitations.
// (GET /trips/{tripId}/confirm)
func (api *Api) GetTripsTripIDConfirm(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Invite someone to the trip.
// (POST /trips/{tripId}/invites)
func (api *Api) PostTripsTripIDInvites(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Get a trip links.
// (GET /trips/{tripId}/links)
func (api *Api) GetTripsTripIDLinks(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Create a trip link.
// (POST /trips/{tripId}/links)
func (api *Api) PostTripsTripIDLinks(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Get a trip participants.
// (GET /trips/{tripId}/participants)
func (api *Api) GetTripsTripIDParticipants(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}
