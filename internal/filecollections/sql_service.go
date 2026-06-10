package filecollections

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mini-cloud/internal/models"
	"mini-cloud/internal/querybuilder"
	"time"

	"github.com/google/uuid"
)

type SQLService struct {
	db *sql.DB
}

func NewSQLService(db *sql.DB) *SQLService {
	return &SQLService{db: db}
}

func (s *SQLService) List() ([]models.FileCollection, error) {
	ctx := context.Background()
	query, args := querybuilder.Select("id", "name", "created_at").
		From("collections").
		OrderBy("created_at").
		SQL()

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	defer rows.Close()

	var collections []models.FileCollection
	for rows.Next() {
		collection, err := scanCollection(rows)
		if err != nil {
			return nil, err
		}

		collection.FileIDs, err = s.listFileIDs(ctx, collection.ID)
		if err != nil {
			return nil, err
		}
		collections = append(collections, collection)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate collections: %w", err)
	}

	return collections, nil
}

func (s *SQLService) Create(name string, fileIDs []models.FileID) (models.FileCollection, error) {
	collection := models.FileCollection{
		BaseModel: models.BaseModel{
			ID:        models.CollectionID(uuid.New().String()),
			CreatedAt: time.Now().UTC(),
		},
		Name:    name,
		FileIDs: normalizeFileIDs(fileIDs),
	}
	if collection.Name == "" {
		return models.FileCollection{}, fmt.Errorf("collection name is required")
	}

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.FileCollection{}, fmt.Errorf("begin create collection transaction: %w", err)
	}
	defer tx.Rollback()

	query, args := querybuilder.InsertInto("collections").
		Columns("id", "name", "created_at").
		Values(string(collection.ID), collection.Name, collection.CreatedAt.Format(time.RFC3339Nano)).
		SQL()
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return models.FileCollection{}, fmt.Errorf("insert collection: %w", err)
	}

	for position, fileID := range collection.FileIDs {
		if err := insertFileID(ctx, tx, collection.ID, fileID, position); err != nil {
			return models.FileCollection{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return models.FileCollection{}, fmt.Errorf("commit create collection transaction: %w", err)
	}

	return collection, nil
}

func (s *SQLService) Get(id models.CollectionID) (models.FileCollection, error) {
	ctx := context.Background()
	collection, err := s.getCollection(ctx, id)
	if err != nil {
		return models.FileCollection{}, err
	}

	collection.FileIDs, err = s.listFileIDs(ctx, collection.ID)
	if err != nil {
		return models.FileCollection{}, err
	}

	return collection, nil
}

func (s *SQLService) AddFile(collectionID models.CollectionID, fileID models.FileID) (models.FileCollection, error) {
	ctx := context.Background()
	collection, err := s.Get(collectionID)
	if err != nil {
		return models.FileCollection{}, err
	}

	normalizedFileIDs := normalizeFileIDs([]models.FileID{fileID})
	if len(normalizedFileIDs) == 0 {
		return models.FileCollection{}, fmt.Errorf("file id is required")
	}
	fileID = normalizedFileIDs[0]
	query, args := querybuilder.Select("COUNT(*)").
		From("collection_files").
		Where("collection_id = ?", string(collectionID)).
		SQL()
	var position int
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(&position); err != nil {
		return models.FileCollection{}, fmt.Errorf("count collection files: %w", err)
	}

	if err := insertFileID(ctx, s.db, collectionID, fileID, position); err != nil {
		return models.FileCollection{}, err
	}

	collection.FileIDs = normalizeFileIDs(append(collection.FileIDs, fileID))
	return collection, nil
}

func (s *SQLService) RemoveFile(collectionID models.CollectionID, fileID models.FileID) (models.FileCollection, error) {
	ctx := context.Background()
	if _, err := s.getCollection(ctx, collectionID); err != nil {
		return models.FileCollection{}, err
	}

	query, args := querybuilder.DeleteFrom("collection_files").
		Where("collection_id = ?", string(collectionID)).
		Where("file_id = ?", string(fileID)).
		SQL()
	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return models.FileCollection{}, fmt.Errorf("delete collection file: %w", err)
	}

	return s.Get(collectionID)
}

func (s *SQLService) Delete(id models.CollectionID) error {
	ctx := context.Background()
	if _, err := s.getCollection(ctx, id); err != nil {
		return err
	}

	query, args := querybuilder.DeleteFrom("collection_files").
		Where("collection_id = ?", string(id)).
		SQL()
	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("delete collection files: %w", err)
	}

	query, args = querybuilder.DeleteFrom("collections").
		Where("id = ?", string(id)).
		SQL()
	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("delete collection: %w", err)
	}

	return nil
}

func (s *SQLService) getCollection(ctx context.Context, id models.CollectionID) (models.FileCollection, error) {
	query, args := querybuilder.Select("id", "name", "created_at").
		From("collections").
		Where("id = ?", string(id)).
		SQL()

	row := s.db.QueryRowContext(ctx, query, args...)
	collection, err := scanCollection(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.FileCollection{}, ErrCollectionNotFound
		}
		return models.FileCollection{}, err
	}

	return collection, nil
}

func (s *SQLService) listFileIDs(ctx context.Context, collectionID models.CollectionID) ([]models.FileID, error) {
	query, args := querybuilder.Select("file_id").
		From("collection_files").
		Where("collection_id = ?", string(collectionID)).
		OrderBy("position").
		SQL()

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list collection files: %w", err)
	}
	defer rows.Close()

	var fileIDs []models.FileID
	for rows.Next() {
		var fileID string
		if err := rows.Scan(&fileID); err != nil {
			return nil, fmt.Errorf("scan collection file: %w", err)
		}
		fileIDs = append(fileIDs, models.FileID(fileID))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate collection files: %w", err)
	}

	return fileIDs, nil
}

type collectionScanner interface {
	Scan(dest ...any) error
}

func scanCollection(scanner collectionScanner) (models.FileCollection, error) {
	var (
		id        string
		name      string
		createdAt string
	)
	if err := scanner.Scan(&id, &name, &createdAt); err != nil {
		return models.FileCollection{}, fmt.Errorf("scan collection: %w", err)
	}

	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return models.FileCollection{}, fmt.Errorf("parse collection created_at: %w", err)
	}

	return models.FileCollection{
		BaseModel: models.BaseModel{
			ID:        models.CollectionID(id),
			CreatedAt: parsedCreatedAt,
		},
		Name: name,
	}, nil
}

type sqlExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func insertFileID(ctx context.Context, executor sqlExecutor, collectionID models.CollectionID, fileID models.FileID, position int) error {
	query, args := querybuilder.InsertInto("collection_files").
		Columns("collection_id", "file_id", "position").
		Values(string(collectionID), string(fileID), position).
		SQL()
	if _, err := executor.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("insert collection file: %w", err)
	}
	return nil
}
