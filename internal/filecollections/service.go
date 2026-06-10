package filecollections

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mini-cloud/internal/models"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	metadataPrefix = "__file_collections/"
	indexKey       = metadataPrefix + "index.json"
)

var ErrCollectionNotFound = NotFoundError{Entity: "file collection"}

type NotFoundError struct {
	Entity string
}

func (e NotFoundError) Error() string {
	return e.Entity + " not found"
}

func (e NotFoundError) Is(target error) bool {
	_, ok := target.(NotFoundError)
	return ok
}

type objectStore interface {
	Save(id string, r io.Reader) error
	Get(id string) (io.ReadCloser, error)
	Delete(id string) error
	List() ([]string, error)
}

type Service struct {
	store objectStore
}

func NewService(store objectStore) *Service {
	return &Service{store: store}
}

func IsMetadataKey(key string) bool {
	return strings.HasPrefix(key, metadataPrefix)
}

func (s *Service) List() ([]models.FileCollection, error) {
	index, err := s.readIndex()
	if err != nil {
		return nil, err
	}

	collections := make([]models.FileCollection, 0, len(index.IDs))
	for _, id := range index.IDs {
		collection, err := s.Get(id)
		if err != nil {
			return nil, err
		}
		collections = append(collections, collection)
	}

	return collections, nil
}

func (s *Service) Create(name string, fileIDs []models.FileID) (models.FileCollection, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.FileCollection{}, fmt.Errorf("collection name is required")
	}

	index, err := s.readIndex()
	if err != nil {
		return models.FileCollection{}, err
	}

	collection := models.FileCollection{
		BaseModel: models.BaseModel{
			ID:        models.CollectionID(uuid.New().String()),
			CreatedAt: time.Now().UTC(),
		},
		Name:    name,
		FileIDs: normalizeFileIDs(fileIDs),
	}

	if err := s.saveCollection(collection); err != nil {
		return models.FileCollection{}, err
	}

	index.IDs = append(index.IDs, collection.ID)
	if err := s.writeIndex(index); err != nil {
		return models.FileCollection{}, err
	}

	return collection, nil
}

func (s *Service) Get(id models.CollectionID) (models.FileCollection, error) {
	id = models.CollectionID(strings.TrimSpace(string(id)))
	if id == "" {
		return models.FileCollection{}, ErrCollectionNotFound
	}

	exists, err := s.objectExists(collectionKey(id))
	if err != nil {
		return models.FileCollection{}, err
	}
	if !exists {
		return models.FileCollection{}, ErrCollectionNotFound
	}

	reader, err := s.store.Get(collectionKey(id))
	if err != nil {
		return models.FileCollection{}, ErrCollectionNotFound
	}
	defer reader.Close()

	var collection models.FileCollection
	if err := json.NewDecoder(reader).Decode(&collection); err != nil {
		return models.FileCollection{}, fmt.Errorf("failed to decode collection: %w", err)
	}

	return collection, nil
}

func (s *Service) AddFile(collectionID models.CollectionID, fileID models.FileID) (models.FileCollection, error) {
	fileID = models.FileID(strings.TrimSpace(string(fileID)))
	if fileID == "" {
		return models.FileCollection{}, fmt.Errorf("file id is required")
	}

	collection, err := s.Get(collectionID)
	if err != nil {
		return models.FileCollection{}, err
	}

	collection.FileIDs = normalizeFileIDs(append(collection.FileIDs, fileID))
	if err := s.saveCollection(collection); err != nil {
		return models.FileCollection{}, err
	}

	return collection, nil
}

func (s *Service) RemoveFile(collectionID models.CollectionID, fileID models.FileID) (models.FileCollection, error) {
	fileID = models.FileID(strings.TrimSpace(string(fileID)))
	if fileID == "" {
		return models.FileCollection{}, fmt.Errorf("file id is required")
	}

	collection, err := s.Get(collectionID)
	if err != nil {
		return models.FileCollection{}, err
	}

	collection.FileIDs = removeFileID(collection.FileIDs, fileID)
	if err := s.saveCollection(collection); err != nil {
		return models.FileCollection{}, err
	}

	return collection, nil
}

func (s *Service) Delete(id models.CollectionID) error {
	id = models.CollectionID(strings.TrimSpace(string(id)))
	if id == "" {
		return ErrCollectionNotFound
	}

	exists, err := s.objectExists(collectionKey(id))
	if err != nil {
		return err
	}
	if !exists {
		return ErrCollectionNotFound
	}

	if err := s.store.Delete(collectionKey(id)); err != nil {
		return err
	}

	index, err := s.readIndex()
	if err != nil {
		return err
	}
	index.IDs = removeCollectionID(index.IDs, id)
	return s.writeIndex(index)
}

type collectionIndex struct {
	IDs []models.CollectionID `json:"ids"`
}

func (s *Service) readIndex() (collectionIndex, error) {
	exists, err := s.objectExists(indexKey)
	if err != nil {
		return collectionIndex{}, err
	}
	if !exists {
		return collectionIndex{}, nil
	}

	reader, err := s.store.Get(indexKey)
	if err != nil {
		return collectionIndex{}, err
	}
	defer reader.Close()

	var index collectionIndex
	if err := json.NewDecoder(reader).Decode(&index); err != nil && !errors.Is(err, io.EOF) {
		return collectionIndex{}, fmt.Errorf("failed to decode collection index: %w", err)
	}

	return index, nil
}

func (s *Service) writeIndex(index collectionIndex) error {
	data, err := json.Marshal(index)
	if err != nil {
		return err
	}
	return s.store.Save(indexKey, bytes.NewReader(data))
}

func (s *Service) saveCollection(collection models.FileCollection) error {
	data, err := json.Marshal(collection)
	if err != nil {
		return err
	}
	return s.store.Save(collectionKey(collection.ID), bytes.NewReader(data))
}

func (s *Service) objectExists(key string) (bool, error) {
	keys, err := s.store.List()
	if err != nil {
		return false, err
	}

	if slices.Contains(keys, key) {
		return true, nil
	}

	return false, nil
}

func collectionKey(id models.CollectionID) string {
	return metadataPrefix + string(id) + ".json"
}

func normalizeFileIDs(fileIDs []models.FileID) []models.FileID {
	seen := make(map[models.FileID]struct{}, len(fileIDs))
	result := make([]models.FileID, 0, len(fileIDs))

	for _, fileID := range fileIDs {
		fileID = models.FileID(strings.TrimSpace(string(fileID)))
		if fileID == "" {
			continue
		}
		if _, ok := seen[fileID]; ok {
			continue
		}
		seen[fileID] = struct{}{}
		result = append(result, fileID)
	}

	return result
}

func removeFileID(fileIDs []models.FileID, fileID models.FileID) []models.FileID {
	result := fileIDs[:0]
	for _, existingFileID := range fileIDs {
		if existingFileID != fileID {
			result = append(result, existingFileID)
		}
	}
	return result
}

func removeCollectionID(collectionIDs []models.CollectionID, collectionID models.CollectionID) []models.CollectionID {
	result := collectionIDs[:0]
	for _, existingCollectionID := range collectionIDs {
		if existingCollectionID != collectionID {
			result = append(result, existingCollectionID)
		}
	}
	return result
}
