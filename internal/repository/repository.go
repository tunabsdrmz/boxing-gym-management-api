package repository

import (
	"context"
	"database/sql"
	"time"
)

type Repository struct {
	User interface {
		CreateUser(ctx context.Context, req CreateUserRequest) (User, error)
		GetUserByEmail(ctx context.Context, email string) (UserWithSecret, error)
		GetUserByID(ctx context.Context, id string) (User, error)
		ListUsers(ctx context.Context, limit, offset int) (ListUsersResult, error)
		UpdateUserAdmin(ctx context.Context, req UpdateUserAdminRequest) (User, error)
		SetPasswordHash(ctx context.Context, userID, passwordHash string) error
	}
	AuthToken interface {
		InsertRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
		ValidateRefreshToken(ctx context.Context, tokenHash string) (userID string, err error)
		DeleteRefreshToken(ctx context.Context, tokenHash string) error
		RevokeRefreshForUser(ctx context.Context, userID string) error
		InsertPasswordResetToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
		ConsumePasswordResetToken(ctx context.Context, tokenHash string) (userID string, err error)
	}
	Fighter interface {
		CreateFighter(ctx context.Context, req CreateFighterRequest) (Fighter, error)
		GetFighterByID(ctx context.Context, req GetFighterRequest) (Fighter, error)
		GetAllFighters(ctx context.Context, req GetAllFightersRequest) (FighterListResult, error)
		UpdateFighter(ctx context.Context, req UpdateFighterRequest) (Fighter, error)
		DeleteFighter(ctx context.Context, req DeleteFighterRequest) error
	}
	FighterAssistant interface {
		ListByFighterID(ctx context.Context, fighterID string) ([]FighterAssistantTrainer, error)
		ListByFighterIDs(ctx context.Context, fighterIDs []string) (map[string][]FighterAssistantTrainer, error)
		Add(ctx context.Context, fighterID, trainerID, role string) error
		Remove(ctx context.Context, fighterID, trainerID string) error
		PrimaryTrainerID(ctx context.Context, fighterID string) (string, error)
	}
	Trainer interface {
		CreateTrainer(ctx context.Context, req CreateTrainerRequest) (Trainer, error)
		GetTrainerByID(ctx context.Context, req GetTrainerRequest) (Trainer, error)
		GetAllTrainers(ctx context.Context, req GetAllTrainersRequest) (TrainerListResult, error)
		UpdateTrainer(ctx context.Context, req UpdateTrainerRequest) (Trainer, error)
		DeleteTrainer(ctx context.Context, req DeleteTrainerRequest) error
	}
	Schedule interface {
		HasOverlap(ctx context.Context, excludeID *string, resourceType string, start, end time.Time) (bool, error)
		Create(ctx context.Context, req CreateScheduleEventRequest) (ScheduleEvent, error)
		GetByID(ctx context.Context, id string) (ScheduleEvent, error)
		List(ctx context.Context, req ListScheduleRequest) ([]ScheduleEvent, error)
		Update(ctx context.Context, req UpdateScheduleEventRequest) (ScheduleEvent, error)
		Delete(ctx context.Context, id string) error
	}
	Attendance interface {
		Upsert(ctx context.Context, req UpsertAttendanceRequest) (AttendanceRecord, error)
		ListByDate(ctx context.Context, gymDate string) ([]AttendanceRecord, error)
		Delete(ctx context.Context, id string) error
	}
	Announcement interface {
		Create(ctx context.Context, req CreateAnnouncementRequest) (Announcement, error)
		ListActive(ctx context.Context) ([]Announcement, error)
		ListAll(ctx context.Context) ([]Announcement, error)
		Update(ctx context.Context, req UpdateAnnouncementRequest) (Announcement, error) // ClearExpires clears expiry
		Delete(ctx context.Context, id string) error
	}
}

func NewRepository(db *sql.DB) Repository {
	return Repository{
		User:             &UserRepository{db: db},
		AuthToken:        &AuthTokenRepository{db: db},
		Fighter:          &FighterRepository{db: db},
		FighterAssistant: &FighterAssistantRepository{db: db},
		Trainer:          &TrainerRepository{db: db},
		Schedule:         &ScheduleRepository{db: db},
		Attendance:       &AttendanceRepository{db: db},
		Announcement:     &AnnouncementRepository{db: db},
	}
}
