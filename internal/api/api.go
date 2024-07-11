package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"nlw-go-course/internal/api/spec"
	"nlw-go-course/internal/pgstore"

	"github.com/discord-gophers/goapi-gen/types"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type store interface{
	CreateTrip(ctx context.Context, pool *pgxpool.Pool, params spec.NewTripBody) (uuid.UUID, error)
	GetTrip(ctx context.Context, id uuid.UUID) (pgstore.Trip, error)
	UpdateTrip(ctx context.Context, arg pgstore.UpdateTripParams) error
	GetParticipants(ctx context.Context, tripID uuid.UUID) ([]pgstore.Participant, error)
	GetParticipant(ctx context.Context, participantID uuid.UUID) (pgstore.Participant, error)
	CreateParticipant(ctx context.Context, arg pgstore.CreateParticipantParams) error
	ConfirmParticipant(ctx context.Context, participantID uuid.UUID) error
	CreateActivity(ctx context.Context, arg pgstore.CreateActivityParams) (uuid.UUID, error)
	GetTripActivities(ctx context.Context, tripID uuid.UUID) ([]pgstore.Activity, error)
	CreateTripLink(ctx context.Context, arg pgstore.CreateTripLinkParams) (uuid.UUID, error)
	GetTripLinks(ctx context.Context, tripID uuid.UUID) ([]pgstore.Link, error)
}

type mailer interface {
	SendConfirmTripEmailToTripOwner(tripID uuid.UUID) error
}

type Api struct{
	store store
	logger *zap.Logger
	validator *validator.Validate
	pool *pgxpool.Pool
	mailer mailer
} 

func NewApi(pool *pgxpool.Pool, logger *zap.Logger, mailer mailer) Api {
	validator := validator.New(validator.WithRequiredStructEnabled())
	return Api {pgstore.New(pool), logger, validator, pool, mailer}
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
	var body spec.NewTripBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "Invalid JSON: " + err.Error()})
	}

	if err := api.validator.Struct(body); err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "Invalid  input: " + err.Error()})
	}
	
	tripID, err := api.store.CreateTrip(r.Context(), api.pool, body)
	if err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "Failed to create trip, try again later"})
	}

	go func () {
		if err := api.mailer.SendConfirmTripEmailToTripOwner(tripID); err != nil {
			api.logger.Error("Failed to send email on PostTrips: %w", zap.Error(err), zap.String("trip_id", tripID.String()))
		}
	}()

	return spec.PostTripsJSON201Response(spec.CreatedNewTrip{TripID: tripID.String()})

}

// Get a trip details.
// (GET /trips/{tripId})
func (api *Api) GetTripsTripID(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDJSON400Response(spec.Error{Message: "Invalid uuid"})
	}

	trip, err := api.store.GetTrip(r.Context(),id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows){
			return spec.GetTripsTripIDJSON400Response(spec.Error{Message: "Trip not found"})
		}
		api.logger.Error("Failed to get trip details", zap.Error(err), zap.String("trip_id",tripID))
		return spec.GetTripsTripIDJSON400Response(spec.Error{Message: "Something went wrong, try again later"})
	}
	
	tripDetails := spec.GetTripDetailsInner{
		ID: trip.ID.String(), Destination: trip.Destination, 
		StartsAt: trip.StartsAt.Time, EndsAt: trip.EndsAt.Time, 
		IsConfirmed: trip.IsConfirmed,
	}

	return spec.GetTripsTripIDJSON200Response(spec.GetTripDetailsResponse{Trip: tripDetails})
}

// Update a trip.
// (PUT /trips/{tripId})
func (api *Api) PutTripsTripID(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "Invalid uuid"})
	}

	_, err = api.store.GetTrip(r.Context(), id)
	if err != nil {
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "Failed to get trip: " + err.Error()})
	}

	var body pgstore.UpdateTripParams
	body.ID = id
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "Invalid JSON: " + err.Error()})
	}

	if err := api.validator.Struct(body); err != nil {
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "Invalid input: " + err.Error()})
	}

	if err := api.store.UpdateTrip(r.Context(), body); err != nil {
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "Failed to update trip, try again later"})
	}

	return spec.PutTripsTripIDJSON204Response(nil)
}

// Get a trip activities.
// (GET /trips/{tripId}/activities)
func (api *Api) GetTripsTripIDActivities(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	panic("not implemented") // TODO: Implement
}

// Create a trip activity.
// (POST /trips/{tripId}/activities)
func (api *Api) PostTripsTripIDActivities(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PostTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Invalid uuid"})
	}

	_, err = api.store.GetTrip(r.Context(), id)
	if err != nil {
		return spec.PostTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Failed to get trip: " + err.Error()})
	}

	var body pgstore.CreateActivityParams
	body.TripID = id
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return spec.PostTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Invalid JSON: " + err.Error()})
	}

	if err := api.validator.Struct(body); err != nil {
		return spec.PostTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Invalid input: " + err.Error()})
	}


	activityId, err := api.store.CreateActivity(r.Context(), body); 
	if err != nil {
		return spec.PostTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Failed to create activity, try again later" + err.Error()})
	}

	return spec.PostTripsTripIDActivitiesJSON201Response(spec.CreatedTripActivities{ActivityID: activityId.String()})
}

// Confirm a trip and send e-mail invitations.
// (GET /trips/{tripId}/confirm)
func (api *Api) GetTripsTripIDConfirm(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "Invalid uuid"})
	}

	trip, err := api.store.GetTrip(r.Context(), id)
	if err != nil {
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "Failed to get trip: " + err.Error()})
	}

	if trip.IsConfirmed {
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "Trip is already confirmed"})
	}

	confirmedTrip := pgstore.UpdateTripParams{
		ID: trip.ID,
		IsConfirmed: true,
		Destination: trip.Destination,
		StartsAt: trip.StartsAt,
		EndsAt: trip.EndsAt,
	}
	
	if err := api.store.UpdateTrip(r.Context(), confirmedTrip); err != nil {
		api.logger.Error("Failed to confirm trip", zap.Error(err), zap.String("trip_id",tripID))
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "Something went wrong, try again later"})
	}

	return spec.GetTripsTripIDConfirmJSON204Response(nil)
}

// Invite someone to the trip.
// (POST /trips/{tripId}/invites)
func (api *Api) PostTripsTripIDInvites(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "Invalid uuid"})
	}

	_, err = api.store.GetTrip(r.Context(), id)
	if err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "Failed to get trip: " + err.Error()})
	}

	var body pgstore.CreateParticipantParams
	body.TripID = id
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "Invalid JSON: " + err.Error()})
	}

	if err := api.validator.Struct(body); err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "Invalid input: " + err.Error()})
	}


	err = api.store.CreateParticipant(r.Context(), body); 
	if err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "Failed to create participant, try again later" + err.Error()})
	}

	return spec.PostTripsTripIDInvitesJSON201Response(nil)
}

// Get a trip links.
// (GET /trips/{tripId}/links)
func (api *Api) GetTripsTripIDLinks(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDLinksJSON400Response(spec.Error{Message: "Invalid uuid"})
	}

	_, err = api.store.GetTrip(r.Context(),id)
	if err != nil {
		return spec.GetTripsTripIDLinksJSON400Response(spec.Error{Message: "Failed to get trip: " + err.Error()})
	}

	tripLinks, err := api.store.GetTripLinks(r.Context(), id)
	if err != nil {
		return spec.GetTripsTripIDLinksJSON400Response(spec.Error{Message: "Failed to get trip links: " + err.Error()})
	}

	
	links := make([]spec.GetTripLinksInner, len(tripLinks))
	for i, eti := range tripLinks{
		links[i] = spec.GetTripLinksInner{
			 ID: eti.ID.String(),
			 Title: eti.Title,
			 URL: eti.Url,
		}
	}

	return spec.GetTripsTripIDLinksJSON200Response(spec.GetTripLinksResponse{Links: links})
}

// Create a trip link.
// (POST /trips/{tripId}/links)
func (api *Api) PostTripsTripIDLinks(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PostTripsTripIDLinksJSON400Response(spec.Error{Message: "Invalid uuid"})
	}

	_, err = api.store.GetTrip(r.Context(), id)
	if err != nil {
		return spec.PostTripsTripIDLinksJSON400Response(spec.Error{Message: "Failed to get trip: " + err.Error()})
	}

	var body pgstore.CreateTripLinkParams
	body.TripID = id
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return spec.PostTripsTripIDLinksJSON400Response(spec.Error{Message: "Invalid JSON: " + err.Error()})
	}

	if err := api.validator.Struct(body); err != nil {
		return spec.PostTripsTripIDLinksJSON400Response(spec.Error{Message: "Invalid input: " + err.Error()})
	}


	linkId, err := api.store.CreateTripLink(r.Context(), body); 
	if err != nil {
		return spec.PostTripsTripIDLinksJSON400Response(spec.Error{Message: "Failed to create link, try again later" + err.Error()})
	}

	return spec.PostTripsTripIDLinksJSON201Response(spec.CreatedTripLinks{LinkID: linkId.String()})
}

// Get a trip participants.
// (GET /trips/{tripId}/participants)
func (api *Api) GetTripsTripIDParticipants(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDParticipantsJSON400Response(spec.Error{Message: "Invalid uuid"})
	}

	_, err = api.store.GetTrip(r.Context(),id)
	if err != nil {
		return spec.GetTripsTripIDParticipantsJSON400Response(spec.Error{Message: "Failed to get trip: " + err.Error()})
	}

	tripParticipants, err := api.store.GetParticipants(r.Context(), id)
	if err != nil {
		return spec.GetTripsTripIDParticipantsJSON400Response(spec.Error{Message: "Failed to get trip participants: " + err.Error()})
	}

	participants := make([]spec.GetTripParticipantsInner, len(tripParticipants))
	for i, eti := range tripParticipants{
		participants[i] = spec.GetTripParticipantsInner{
			 ID: eti.ID.String(),
			 IsConfirmed: eti.IsConfirmed,
			 Email: types.Email(eti.Email),
			 Name: nil,
		}
	}

	return spec.GetTripsTripIDParticipantsJSON200Response(spec.GetTripParticipantsResponse{Participants: participants})
}
