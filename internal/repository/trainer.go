package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)


type Trainer struct {
	ID string `json:"id" default:"uuid_generate_v4()"`
	Name string `json:"name"`
	Age int `json:"age"`
	Specialization []string `json:"specialization"`
	FighterIds []string `json:"fighter_ids"`
	Fighters []Fighter `json:"fighters"`
	CreatedAt time.Time `json:"created_at" default:"now()"`
	UpdatedAt time.Time `json:"updated_at" default:"now()"`
}

type TrainerRepository struct {
	db *sql.DB
}

type CreateTrainerRequest struct {
	Name string `json:"name"`
	Age int `json:"age"`
	Specialization []string `json:"specialization"`
}

func (t *TrainerRepository) CreateTrainer(ctx context.Context, req CreateTrainerRequest) (Trainer, error) {
	query := `
		INSERT INTO trainers (name, age, specialization)
		VALUES ($1, $2, $3)
		RETURNING id, name, age, specialization, created_at, updated_at
		`
		var trainer Trainer
		row := t.db.QueryRowContext(ctx, query, req.Name, req.Age, pq.Array(req.Specialization))
		err := row.Scan(&trainer.ID, &trainer.Name, &trainer.Age, pq.Array(&trainer.Specialization), &trainer.CreatedAt, &trainer.UpdatedAt)
		if err != nil {
			return Trainer{}, err
		}
		if err := t.attachFighters(ctx, &trainer); err != nil {
			return Trainer{}, err
		}
		return trainer, nil
	}

type GetTrainerRequest struct {
	ID string `json:"id"`
}

func (t *TrainerRepository) GetTrainerByID(ctx context.Context, req GetTrainerRequest) (Trainer, error) {
	query := `
		SELECT id, name, age, specialization, created_at, updated_at
		FROM trainers
		WHERE id = $1
	`
	var trainer Trainer
	row := t.db.QueryRowContext(ctx, query, req.ID)
	err := row.Scan(&trainer.ID, &trainer.Name, &trainer.Age, pq.Array(&trainer.Specialization), &trainer.CreatedAt, &trainer.UpdatedAt)
	if err != nil {
		return Trainer{}, err
	}
	if err := t.attachFighters(ctx, &trainer); err != nil {
		return Trainer{}, err
	}
	return trainer, nil
}

type GetAllTrainersRequest struct {
	Limit string `json:"limit"`
	Offset string `json:"offset"`
}

// TrainerListResult is the trainers page plus total row count for pagination.
type TrainerListResult struct {
	Trainers []Trainer
	Total    int
}

func (t *TrainerRepository) GetAllTrainers(ctx context.Context, req GetAllTrainersRequest) (TrainerListResult, error) {
	var total int
	countRow := t.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM trainers`)
	if err := countRow.Scan(&total); err != nil {
		return TrainerListResult{}, err
	}

	query := `
		SELECT id, name, age, specialization, created_at, updated_at
		FROM trainers
		ORDER BY created_at DESC
		LIMIT $1
		OFFSET $2
	`
	rows, err := t.db.QueryContext(ctx, query, req.Limit, req.Offset)
	if err != nil {
		return TrainerListResult{}, err
	}
	defer rows.Close()
	var trainers []Trainer
	for rows.Next() {
		var trainer Trainer
		err := rows.Scan(&trainer.ID, &trainer.Name, &trainer.Age, pq.Array(&trainer.Specialization), &trainer.CreatedAt, &trainer.UpdatedAt)
		if err != nil {
			return TrainerListResult{}, err
		}
		if err := t.attachFighters(ctx, &trainer); err != nil {
			return TrainerListResult{}, err
		}
		trainers = append(trainers, trainer)
	}
	if err := rows.Err(); err != nil {
		return TrainerListResult{}, err
	}
	return TrainerListResult{Trainers: trainers, Total: total}, nil
}

func (t *TrainerRepository) attachFighters(ctx context.Context, trainer *Trainer) error {
	fighters, err := t.fightersByTrainerID(ctx, trainer.ID)
	if err != nil {
		return err
	}
	trainer.Fighters = fighters
	trainer.FighterIds = make([]string, len(fighters))
	for i := range fighters {
		trainer.FighterIds[i] = fighters[i].ID
	}
	return nil
}

func (t *TrainerRepository) fightersByTrainerID(ctx context.Context, trainerID string) ([]Fighter, error) {
	query := `
		SELECT id, name, age, weight, wins, losses, trainer_id, created_at, updated_at
		FROM fighters
		WHERE trainer_id = $1
		ORDER BY created_at DESC
	`
	rows, err := t.db.QueryContext(ctx, query, trainerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var fighters []Fighter
	for rows.Next() {
		var f Fighter
		err := rows.Scan(&f.ID, &f.Name, &f.Age, &f.Weight, &f.Wins, &f.Losses, &f.TrainerID, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, err
		}
		fighters = append(fighters, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return fighters, nil
}

type UpdateTrainerRequest struct {
	ID string `json:"id"`
	Name *string `json:"name"`
	Age *int `json:"age"`
	Specialization *[]string `json:"specialization"`
}

// UpdateTrainerRequest supports partial updates: only non-nil pointer fields are written.
func (t *TrainerRepository) UpdateTrainer(ctx context.Context, req UpdateTrainerRequest) (Trainer, error) {
	var setParts []string
	var args []any
	n := 1

	if req.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", n))
		args = append(args, *req.Name)
		n++
	}
	if req.Age != nil {
		setParts = append(setParts, fmt.Sprintf("age = $%d", n))
		args = append(args, *req.Age)
		n++
	}
	if req.Specialization != nil {
		setParts = append(setParts, fmt.Sprintf("specialization = $%d", n))
		args = append(args, pq.Array(*req.Specialization))
		n++
	}

	setParts = append(setParts, "updated_at = now()")

	query := fmt.Sprintf(`
		UPDATE trainers
		SET %s
		WHERE id = $%d
		RETURNING id, name, age, specialization, created_at, updated_at
	`, strings.Join(setParts, ", "), n)
	args = append(args, req.ID)

	var trainer Trainer
	row := t.db.QueryRowContext(ctx, query, args...)
	err := row.Scan(&trainer.ID, &trainer.Name, &trainer.Age, pq.Array(&trainer.Specialization), &trainer.CreatedAt, &trainer.UpdatedAt)
	if err != nil {
		return Trainer{}, err
	}
	if err := t.attachFighters(ctx, &trainer); err != nil {
		return Trainer{}, err
	}
	return trainer, nil
}

type DeleteTrainerRequest struct {
	ID string `json:"id"`
}

func (t *TrainerRepository) DeleteTrainer(ctx context.Context, req DeleteTrainerRequest) error {
	query := `
		DELETE FROM trainers
		WHERE id = $1
	`
	_, err := t.db.ExecContext(ctx, query, req.ID)
	if err != nil {
		return err
	}
	return nil
}