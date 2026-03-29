package repository

import (
	"context"
	"database/sql"
)


type Repository struct {
	User interface {
		CreateUser(ctx context.Context, req CreateUserRequest) (User, error)
		GetUserByEmail(ctx context.Context, email string) (UserWithSecret, error)
	}
	Fighter interface{
		CreateFighter(ctx context.Context, req CreateFighterRequest) (Fighter, error)
		GetFighterByID(ctx context.Context, req GetFighterRequest) (Fighter, error)
		GetAllFighters(ctx context.Context, req GetAllFightersRequest) (FighterListResult, error)
		UpdateFighter(ctx context.Context, req UpdateFighterRequest) (Fighter, error)
		DeleteFighter(ctx context.Context, req DeleteFighterRequest) error
	}
	Trainer interface{
		CreateTrainer(ctx context.Context, req CreateTrainerRequest) (Trainer, error)
		GetTrainerByID(ctx context.Context, req GetTrainerRequest) (Trainer, error)
		GetAllTrainers(ctx context.Context, req GetAllTrainersRequest) (TrainerListResult, error)
		UpdateTrainer(ctx context.Context, req UpdateTrainerRequest) (Trainer, error)
		DeleteTrainer(ctx context.Context, req DeleteTrainerRequest) error
	}
}

func NewRepository(db *sql.DB) Repository {
	return Repository{
		User:    &UserRepository{db: db},
		Fighter: &FighterRepository{db: db},
		Trainer: &TrainerRepository{db: db},
	}
}