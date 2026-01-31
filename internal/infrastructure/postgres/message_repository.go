package postgres

import (
	"context"

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

func (r *MessageRepository) Create(ctx context.Context, m messages.Message) (*messages.Message, error) {
	// Prepare parameters
	params := sqlc.CreateMessageParams{
		SenderID: m.SenderID,
		Content:  m.Content,
		ProjectID: pgtype.Int8{
			Int64: func() int64 {
				if m.ProjectID != nil {
					return *m.ProjectID
				}
				return 0
			}(),
			Valid: m.ProjectID != nil,
		},
		ReceiverID: pgtype.Int8{
			Int64: func() int64 {
				if m.ReceiverID != nil {
					return *m.ReceiverID
				}
				return 0
			}(),
			Valid: m.ReceiverID != nil,
		},
	}

	res, err := r.queries.CreateMessage(ctx, params)
	if err != nil {
		return nil, err
	}

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
		return nil, err
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
		return nil, err
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
		if row.ReceiverID.Valid {
			val := row.ReceiverID.Int64
			msg.ReceiverID = &val
		}
		list = append(list, msg)
	}
	return list, nil
}
