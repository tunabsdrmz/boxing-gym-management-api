package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type FighterAssistantTrainer struct {
	TrainerID string `json:"trainer_id"`
	Role      string `json:"role"`
}

type FighterAssistantRepository struct {
	db *sql.DB
}

func (f *FighterAssistantRepository) ListByFighterID(ctx context.Context, fighterID string) ([]FighterAssistantTrainer, error) {
	rows, err := f.db.QueryContext(ctx, `
		SELECT trainer_id::text, role
		FROM fighter_assistant_trainers
		WHERE fighter_id = $1::uuid
		ORDER BY role, trainer_id
	`, fighterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FighterAssistantTrainer
	for rows.Next() {
		var a FighterAssistantTrainer
		if err := rows.Scan(&a.TrainerID, &a.Role); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (f *FighterAssistantRepository) ListByFighterIDs(ctx context.Context, fighterIDs []string) (map[string][]FighterAssistantTrainer, error) {
	if len(fighterIDs) == 0 {
		return map[string][]FighterAssistantTrainer{}, nil
	}
	ph := make([]string, len(fighterIDs))
	args := make([]any, len(fighterIDs))
	for i, id := range fighterIDs {
		ph[i] = fmt.Sprintf("$%d::uuid", i+1)
		args[i] = id
	}
	q := fmt.Sprintf(`
		SELECT fighter_id::text, trainer_id::text, role
		FROM fighter_assistant_trainers
		WHERE fighter_id IN (%s)
	`, strings.Join(ph, ","))
	rows, err := f.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string][]FighterAssistantTrainer)
	for rows.Next() {
		var fid, tid, role string
		if err := rows.Scan(&fid, &tid, &role); err != nil {
			return nil, err
		}
		m[fid] = append(m[fid], FighterAssistantTrainer{TrainerID: tid, Role: role})
	}
	return m, rows.Err()
}

func (f *FighterAssistantRepository) Add(ctx context.Context, fighterID, trainerID, role string) error {
	if role != "assistant" && role != "corner" {
		return errors.New("invalid assistant role")
	}
	_, err := f.db.ExecContext(ctx, `
		INSERT INTO fighter_assistant_trainers (fighter_id, trainer_id, role)
		VALUES ($1::uuid, $2::uuid, $3)
		ON CONFLICT (fighter_id, trainer_id) DO UPDATE SET role = EXCLUDED.role
	`, fighterID, trainerID, role)
	return err
}

func (f *FighterAssistantRepository) Remove(ctx context.Context, fighterID, trainerID string) error {
	res, err := f.db.ExecContext(ctx, `
		DELETE FROM fighter_assistant_trainers
		WHERE fighter_id = $1::uuid AND trainer_id = $2::uuid
	`, fighterID, trainerID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (f *FighterAssistantRepository) PrimaryTrainerID(ctx context.Context, fighterID string) (string, error) {
	var tid string
	err := f.db.QueryRowContext(ctx, `SELECT trainer_id::text FROM fighters WHERE id = $1::uuid`, fighterID).Scan(&tid)
	return tid, err
}
