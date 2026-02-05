package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nelfander/Playingfield/internal/domain/messages"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

type MessageRepository struct {
	db      *DBAdapter
	queries *sqlc.Queries
}

func NewMessageRepository(db *DBAdapter) *MessageRepository {
	return &MessageRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

// helper function
func int64Value(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func (r *MessageRepository) Create(ctx context.Context, m messages.Message) (*messages.Message, error) {
	// Prepare parameters with clean mapping
	params := sqlc.CreateMessageParams{
		SenderID: m.SenderID,
		Content:  m.Content,
		ProjectID: pgtype.Int8{
			Int64: int64Value(m.ProjectID),
			Valid: m.ProjectID != nil,
		},
		ReceiverID: pgtype.Int8{
			Int64: int64Value(m.ReceiverID),
			Valid: m.ReceiverID != nil,
		},
	}

	res, err := r.queries.CreateMessage(ctx, params)
	if err != nil {
		slog.Error("failed to persist message", "error", err)
		return nil, fmt.Errorf("db: failed to create message: %w", err)
	}

	// debug and not info for messages, in a busy day with info
	// it can fil up pretty fast :p
	// debug wont show in prod
	slog.Debug("message persisted to db", "id", res.ID, "sender_id", res.SenderID)

	// Map back to domain model
	return &messages.Message{
		ID:          res.ID,
		SenderID:    res.SenderID,
		Content:     res.Content,
		CreatedAt:   res.CreatedAt.Time,
		SenderEmail: res.SenderEmail,
		ProjectID:   m.ProjectID,
		ReceiverID:  m.ReceiverID,
	}, nil
}

func (r *MessageRepository) GetByProject(ctx context.Context, projectID int64) ([]messages.Message, error) {
	rows, err := r.queries.GetProjectMessages(ctx, pgtype.Int8{Int64: projectID, Valid: true})
	if err != nil {
		slog.Error("database query failed", "op", "GetByProject", "err", err)
		return nil, fmt.Errorf("db: failed to get project messages: %w", err)
	}

	var list []messages.Message
	for _, row := range rows {
		msg := messages.Message{
			ID:          row.ID,
			SenderID:    row.SenderID,
			Content:     row.Content,
			CreatedAt:   row.CreatedAt.Time,
			SenderEmail: row.SenderEmail,
		}
		if row.ProjectID.Valid {
			val := row.ProjectID.Int64
			msg.ProjectID = &val
		}
		list = append(list, msg)
	}
	return list, nil
}

func (r *MessageRepository) GetDirectMessages(ctx context.Context, userA, userB int64) ([]messages.Message, error) {
	params := sqlc.GetDirectMessagesParams{
		SenderID:   userA,
		ReceiverID: pgtype.Int8{Int64: userB, Valid: true},
	}

	rows, err := r.queries.GetDirectMessages(ctx, params)
	if err != nil {
		slog.Error("database query failed", "op", "GetDirectMessages", "userA", userA, "userB", userB, "err", err)
		return nil, fmt.Errorf("db: failed to get direct messages: %w", err)
	}

	slog.Debug("fetched direct messages", "count", len(rows), "userA", userA, "userB", userB)

	var list []messages.Message
	for _, row := range rows {
		msg := messages.Message{
			ID:          row.ID,
			SenderID:    row.SenderID,
			Content:     row.Content,
			CreatedAt:   row.CreatedAt.Time,
			SenderEmail: row.SenderEmail,
		}
		if row.ReceiverID.Valid {
			val := row.ReceiverID.Int64
			msg.ReceiverID = &val
		}
		list = append(list, msg)
	}
	return list, nil
}
