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
			is_draw, move_number, version, x_wins, o_wins, draws,
			round_number, rematch_x_requested, rematch_o_requested
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
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
		g.XWins,
		g.OWins,
		g.Draws,
		g.RoundNumber,
		g.RematchXRequested,
		g.RematchORequested,
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
			x_wins = $9,
			o_wins = $10,
			draws = $11,
			round_number = $12,
			rematch_x_requested = $13,
			rematch_o_requested = $14,
			updated_at = NOW()
		WHERE id = $1 AND version = $15
	`,
		g.ID,
		statusToString(g.Status),
		boardJSON,
		markToString(g.NextTurn),
		markToString(g.Winner),
		g.IsDraw,
		g.MoveNumber,
		g.Version,
		g.XWins,
		g.OWins,
		g.Draws,
		g.RoundNumber,
		g.RematchXRequested,
		g.RematchORequested,
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
		       is_draw, move_number, version, x_wins, o_wins, draws,
		       round_number, rematch_x_requested, rematch_o_requested
		FROM games
		WHERE %s = $1
	`, column)

	var (
		id                string
		joinCode          string
		status            string
		boardData         []byte
		nextTurn          string
		winner            string
		isDraw            bool
		moveNumber        int64
		version           int64
		xWins             int64
		oWins             int64
		draws             int64
		roundNumber       int64
		rematchXRequested bool
		rematchORequested bool
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
		&xWins,
		&oWins,
		&draws,
		&roundNumber,
		&rematchXRequested,
		&rematchORequested,
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
		ID:                id,
		JoinCode:          joinCode,
		Board:             board,
		Status:            stringToStatus(status),
		NextTurn:          stringToMark(nextTurn),
		Winner:            stringToMark(winner),
		IsDraw:            isDraw,
		MoveNumber:        moveNumber,
		Version:           version,
		XWins:             xWins,
		OWins:             oWins,
		Draws:             draws,
		RoundNumber:       roundNumber,
		RematchXRequested: rematchXRequested,
		RematchORequested: rematchORequested,
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
