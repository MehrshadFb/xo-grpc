package postgres

import (
	"context"
	"errors"
	"fmt"

	domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"
	"github.com/MehrshadFb/xo-grpc/internal/repository"
	"github.com/MehrshadFb/xo-grpc/internal/store/memory"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GameRepository struct {
	pool *pgxpool.Pool
}

func NewGameRepository(pool *pgxpool.Pool) *GameRepository {
	return &GameRepository{
		pool: pool,
	}
}

var _ repository.GameRepository = (*GameRepository)(nil)

func (r *GameRepository) Create(g *domaingame.Game) error {
	boardJSON, err := boardToJSON(g.Board)
	if err != nil {
		return err
	}

	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO games (
			id, join_code, status, board, next_turn, winner,
			is_draw, move_number, version
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`,
		g.ID,
		g.JoinCode,
		statusToString(g.Status),
		boardJSON,
		markToString(g.NextTurn),
		markToString(g.Winner),
		g.IsDraw,
		g.MoveNumber,
		g.Version,
	)
	if err != nil {
		return fmt.Errorf("insert game: %w", err)
	}

	if g.PlayerX != nil {
		if err := insertPlayer(ctx, tx, g.ID, g.PlayerX); err != nil {
			return err
		}
	}

	if g.PlayerO != nil {
		if err := insertPlayer(ctx, tx, g.ID, g.PlayerO); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *GameRepository) GetByID(id string) (*domaingame.Game, error) {
	return r.getByColumn("id", id)
}

func (r *GameRepository) GetByJoinCode(joinCode string) (*domaingame.Game, error) {
	return r.getByColumn("join_code", joinCode)
}

func (r *GameRepository) Update(g *domaingame.Game) error {
	boardJSON, err := boardToJSON(g.Board)
	if err != nil {
		return err
	}

	expectedVersion := g.Version - 1

	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, `
		UPDATE games
		SET
			status = $2,
			board = $3,
			next_turn = $4,
			winner = $5,
			is_draw = $6,
			move_number = $7,
			version = $8,
			updated_at = NOW()
		WHERE id = $1 AND version = $9
	`,
		g.ID,
		statusToString(g.Status),
		boardJSON,
		markToString(g.NextTurn),
		markToString(g.Winner),
		g.IsDraw,
		g.MoveNumber,
		g.Version,
		expectedVersion,
	)
	if err != nil {
		return fmt.Errorf("update game: %w", err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrConflict
	}

	if g.PlayerX != nil {
		if err := upsertPlayer(ctx, tx, g.ID, g.PlayerX); err != nil {
			return err
		}
	}

	if g.PlayerO != nil {
		if err := upsertPlayer(ctx, tx, g.ID, g.PlayerO); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *GameRepository) getByColumn(column string, value string) (*domaingame.Game, error) {
	ctx := context.Background()

	query := fmt.Sprintf(`
		SELECT id, join_code, status, board, next_turn, winner,
		       is_draw, move_number, version
		FROM games
		WHERE %s = $1
	`, column)

	var (
		id         string
		joinCode   string
		status     string
		boardData  []byte
		nextTurn   string
		winner     string
		isDraw     bool
		moveNumber int64
		version    int64
	)

	err := r.pool.QueryRow(ctx, query, value).Scan(
		&id,
		&joinCode,
		&status,
		&boardData,
		&nextTurn,
		&winner,
		&isDraw,
		&moveNumber,
		&version,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, memory.ErrGameNotFound
		}
		return nil, fmt.Errorf("select game: %w", err)
	}

	board, err := jsonToBoard(boardData)
	if err != nil {
		return nil, err
	}

	g := &domaingame.Game{
		ID:         id,
		JoinCode:   joinCode,
		Board:      board,
		Status:     stringToStatus(status),
		NextTurn:   stringToMark(nextTurn),
		Winner:     stringToMark(winner),
		IsDraw:     isDraw,
		MoveNumber: moveNumber,
		Version:    version,
	}

	players, err := r.playersByGameID(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, p := range players {
		switch p.Mark {
		case domaingame.MarkX:
			g.PlayerX = p
		case domaingame.MarkO:
			g.PlayerO = p
		}
	}

	return g, nil
}

func insertPlayer(ctx context.Context, tx pgx.Tx, gameID string, p *domaingame.Player) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO players (id, game_id, display_name, mark)
		VALUES ($1,$2,$3,$4)
	`,
		p.ID,
		gameID,
		p.DisplayName,
		markToString(p.Mark),
	)
	if err != nil {
		return fmt.Errorf("insert player: %w", err)
	}

	return nil
}

func upsertPlayer(ctx context.Context, tx pgx.Tx, gameID string, p *domaingame.Player) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO players (id, game_id, display_name, mark)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (id) DO UPDATE
		SET
			display_name = EXCLUDED.display_name,
			mark = EXCLUDED.mark
	`,
		p.ID,
		gameID,
		p.DisplayName,
		markToString(p.Mark),
	)
	if err != nil {
		return fmt.Errorf("upsert player: %w", err)
	}

	return nil
}

func (r *GameRepository) playersByGameID(ctx context.Context, gameID string) ([]*domaingame.Player, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, display_name, mark
		FROM players
		WHERE game_id = $1
	`, gameID)
	if err != nil {
		return nil, fmt.Errorf("select players: %w", err)
	}
	defer rows.Close()

	var players []*domaingame.Player

	for rows.Next() {
		var (
			id          string
			displayName string
			mark        string
		)

		if err := rows.Scan(&id, &displayName, &mark); err != nil {
			return nil, fmt.Errorf("scan player: %w", err)
		}

		players = append(players, &domaingame.Player{
			ID:          id,
			DisplayName: displayName,
			Mark:        stringToMark(mark),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("players rows: %w", err)
	}

	return players, nil
}
