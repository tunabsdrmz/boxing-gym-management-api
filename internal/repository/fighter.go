package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Fighter struct {
	ID       string `json:"id" default:"uuid_generate_v4()"`
    Name       string `json:"name"`
    Age        int `json:"age"`
    Weight     float64 `json:"weight"`
    Wins       int `json:"wins"`
    Losses     int `json:"losses"`
    TrainerID  string `json:"trainer_id"`
	CreatedAt time.Time `json:"created_at" default:"now()"`
	UpdatedAt time.Time `json:"updated_at" default:"now()"`
}
type FighterRepository struct {
	db *sql.DB
}

type CreateFighterRequest struct {
	Name string `json:"name"`
	Age int `json:"age"`
	Weight float64 `json:"weight"`
	Wins int `json:"wins"`
	Losses int `json:"losses"`
	TrainerID string `json:"trainer_id"`
}

func (f *FighterRepository) CreateFighter(ctx context.Context, req CreateFighterRequest) (Fighter, error) {
	query := `
		INSERT INTO fighters (name, age, weight, wins, losses, trainer_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, age, weight, wins, losses, trainer_id, created_at, updated_at
	`
	var fighter Fighter
	row := f.db.QueryRowContext(ctx, query, req.Name, req.Age, req.Weight, req.Wins, req.Losses, req.TrainerID)
	err := row.Scan(&fighter.ID, &fighter.Name, &fighter.Age, &fighter.Weight, &fighter.Wins, &fighter.Losses, &fighter.TrainerID, &fighter.CreatedAt, &fighter.UpdatedAt)
	if err != nil {
		return Fighter{}, err
	}
	return fighter, nil
}

type GetFighterRequest struct {
	ID string `json:"id"`
}

func (f *FighterRepository) GetFighterByID(ctx context.Context, req GetFighterRequest) (Fighter, error) {
	query := `
		SELECT id, name, age, weight, wins, losses, trainer_id, created_at, updated_at
		FROM fighters
		WHERE id = $1
	`
	var fighter Fighter
	row := f.db.QueryRowContext(ctx, query, req.ID)
	err := row.Scan(&fighter.ID, &fighter.Name, &fighter.Age, &fighter.Weight, &fighter.Wins, &fighter.Losses, &fighter.TrainerID, &fighter.CreatedAt, &fighter.UpdatedAt)
	if err != nil {
		return Fighter{}, err
	}
	return fighter, nil
}

type GetAllFightersRequest struct {
	Limit string `json:"limit"`
	Offset string `json:"offset"`
}

// FighterListResult is the fighters page plus total row count for pagination.
type FighterListResult struct {
	Fighters []Fighter
	Total    int
}

func (f *FighterRepository) GetAllFighters(ctx context.Context, req GetAllFightersRequest) (FighterListResult, error) {
	var total int
	countRow := f.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM fighters`)
	if err := countRow.Scan(&total); err != nil {
		return FighterListResult{}, err
	}

	query := `
		SELECT id, name, age, weight, wins, losses, trainer_id, created_at, updated_at
		FROM fighters
		ORDER BY created_at DESC
		LIMIT $1
		OFFSET $2
	`
	rows, err := f.db.QueryContext(ctx, query, req.Limit, req.Offset)
	if err != nil {
		return FighterListResult{}, err
	}
	defer rows.Close()
	var fighters []Fighter
	for rows.Next() {
		var fighter Fighter
		err := rows.Scan(&fighter.ID, &fighter.Name, &fighter.Age, &fighter.Weight, &fighter.Wins, &fighter.Losses, &fighter.TrainerID, &fighter.CreatedAt, &fighter.UpdatedAt)
		if err != nil {
			return FighterListResult{}, err
		}
		fighters = append(fighters, fighter)
	}
	if err := rows.Err(); err != nil {
		return FighterListResult{}, err
	}
	return FighterListResult{Fighters: fighters, Total: total}, nil
}

// UpdateFighterRequest supports partial updates: only non-nil pointer fields are written.
// Omitted JSON fields stay nil and keep the existing DB values.
type UpdateFighterRequest struct {
	ID        string   `json:"id"`
	Name      *string  `json:"name"`
	Age       *int     `json:"age"`
	Weight    *float64 `json:"weight"`
	Wins      *int     `json:"wins"`
	Losses    *int     `json:"losses"`
	TrainerID *string  `json:"trainer_id"`
}

func (f *FighterRepository) UpdateFighter(ctx context.Context, req UpdateFighterRequest) (Fighter, error) {
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
	if req.Weight != nil {
		setParts = append(setParts, fmt.Sprintf("weight = $%d", n))
		args = append(args, *req.Weight)
		n++
	}
	if req.Wins != nil {
		setParts = append(setParts, fmt.Sprintf("wins = $%d", n))
		args = append(args, *req.Wins)
		n++
	}
	if req.Losses != nil {
		setParts = append(setParts, fmt.Sprintf("losses = $%d", n))
		args = append(args, *req.Losses)
		n++
	}
	if req.TrainerID != nil {
		setParts = append(setParts, fmt.Sprintf("trainer_id = $%d", n))
		args = append(args, *req.TrainerID)
		n++
	}

	setParts = append(setParts, "updated_at = now()")

	query := fmt.Sprintf(`
		UPDATE fighters
		SET %s
		WHERE id = $%d
		RETURNING id, name, age, weight, wins, losses, trainer_id, created_at, updated_at
	`, strings.Join(setParts, ", "), n)
	args = append(args, req.ID)

	var fighter Fighter
	row := f.db.QueryRowContext(ctx, query, args...)
	err := row.Scan(&fighter.ID, &fighter.Name, &fighter.Age, &fighter.Weight, &fighter.Wins, &fighter.Losses, &fighter.TrainerID, &fighter.CreatedAt, &fighter.UpdatedAt)
	if err != nil {
		return Fighter{}, err
	}
	return fighter, nil
}

type DeleteFighterRequest struct {
	ID string `json:"id"`
}

func (f *FighterRepository) DeleteFighter(ctx context.Context, req DeleteFighterRequest) error {
	query := `
		DELETE FROM fighters
		WHERE id = $1
	`
	_, err := f.db.ExecContext(ctx, query, req.ID)
	if err != nil {
		return err
	}
	return nil
}