package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nelfander/Playingfield/internal/domain/messages"
)

type MessageRepository struct {
	db *DBAdapter
}

func NewMessageRepository(db *DBAdapter) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, m messages.Message) (*messages.Message, error) {
	//  pgtype to handle the nullable pointers from domain
	var pID, rID pgtype.Int8

	if m.ProjectID != nil {
		pID = pgtype.Int8{Int64: *m.ProjectID, Valid: true}
	}
	if m.ReceiverID != nil {
		rID = pgtype.Int8{Int64: *m.ReceiverID, Valid: true}
	}

	row := r.db.QueryRow(ctx,
		`INSERT INTO messages (sender_id, content, project_id, receiver_id)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		m.SenderID, m.Content, pID, rID,
	)

	var created messages.Message = m
	var createdAt time.Time

	if err := row.Scan(&created.ID, &createdAt); err != nil {
		return nil, err
	}

	created.CreatedAt = createdAt
	return &created, nil
}

func (r *MessageRepository) GetByProject(ctx context.Context, projectID int64) ([]messages.Message, error) {
	rows, err := r.db.Query(ctx,
		`SELECT m.id, m.sender_id, m.content, m.project_id, m.created_at, u.email
		 FROM messages m
		 JOIN users u ON m.sender_id = u.id
		 WHERE m.project_id = $1
		 ORDER BY m.created_at ASC`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []messages.Message
	for rows.Next() {
		var m messages.Message
		var pID pgtype.Int8
		if err := rows.Scan(&m.ID, &m.SenderID, &m.Content, &pID, &m.CreatedAt, &m.SenderEmail); err != nil {
			return nil, err
		}
		if pID.Valid {
			val := pID.Int64
			m.ProjectID = &val
		}
		list = append(list, m)
	}
	return list, nil
}

func (r *MessageRepository) GetDirectMessages(ctx context.Context, userA, userB int64) ([]messages.Message, error) {
	rows, err := r.db.Query(ctx,
		`SELECT m.id, m.sender_id, m.content, m.receiver_id, m.created_at, u.email
		 FROM messages m
		 JOIN users u ON m.sender_id = u.id
		 WHERE (m.sender_id = $1 AND m.receiver_id = $2)
		    OR (m.sender_id = $2 AND m.receiver_id = $1)
		 ORDER BY m.created_at ASC`,
		userA, userB,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []messages.Message
	for rows.Next() {
		var m messages.Message
		var rID pgtype.Int8
		if err := rows.Scan(&m.ID, &m.SenderID, &m.Content, &rID, &m.CreatedAt, &m.SenderEmail); err != nil {
			return nil, err
		}
		if rID.Valid {
			val := rID.Int64
			m.ReceiverID = &val
		}
		list = append(list, m)
	}
	return list, nil
}
