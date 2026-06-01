package filecollections

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mini-cloud/internal/models"
	"mini-cloud/internal/storage"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	metadataPrefix = "__file_collections/"
	indexKey       = metadataPrefix + "index.json"
)

var ErrCollectionNotFound = errors.New("file collection not found")

type Service interface {
	List() ([]models.FileCollection, error)
	Create(name string, fileIDs []string) (models.FileCollection, error)
	Get(id string) (models.FileCollection, error)
	AddFile(collectionID, fileID string) (models.FileCollection, error)
	RemoveFile(collectionID, fileID string) (models.FileCollection, error)
	Delete(id string) error
}

type service struct {
	store storage.Storage
}

type collectionIndex struct {
	IDs []string `json:"ids"`
}

func NewService(store storage.Storage) Service {
	return &service{store: store}
}

func IsMetadataKey(key string) bool {
	return strings.HasPrefix(key, metadataPrefix)
}

func (s *service) List() ([]models.FileCollection, error) {
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

func (s *service) Create(name string, fileIDs []string) (models.FileCollection, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.FileCollection{}, fmt.Errorf("collection name is required")
	}

	index, err := s.readIndex()
	if err != nil {
		return models.FileCollection{}, err
	}

	collection := models.FileCollection{
		ID:        uuid.New().String(),
		Name:      name,
		FileIDs:   normalizeFileIDs(fileIDs),
		CreatedAt: time.Now().UTC(),
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

func (s *service) Get(id string) (models.FileCollection, error) {
	id = strings.TrimSpace(id)
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

func (s *service) AddFile(collectionID, fileID string) (models.FileCollection, error) {
	fileID = strings.TrimSpace(fileID)
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

func (s *service) RemoveFile(collectionID, fileID string) (models.FileCollection, error) {
	fileID = strings.TrimSpace(fileID)
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

func (s *service) Delete(id string) error {
	id = strings.TrimSpace(id)
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
	index.IDs = removeFileID(index.IDs, id)
	return s.writeIndex(index)
}

func (s *service) readIndex() (collectionIndex, error) {
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

func (s *service) writeIndex(index collectionIndex) error {
	data, err := json.Marshal(index)
	if err != nil {
		return err
	}
	return s.store.Save(indexKey, bytes.NewReader(data))
}

func (s *service) saveCollection(collection models.FileCollection) error {
	data, err := json.Marshal(collection)
	if err != nil {
		return err
	}
	return s.store.Save(collectionKey(collection.ID), bytes.NewReader(data))
}

func (s *service) objectExists(key string) (bool, error) {
	keys, err := s.store.List()
	if err != nil {
		return false, err
	}

	for _, existingKey := range keys {
		if existingKey == key {
			return true, nil
		}
	}

	return false, nil
}

func collectionKey(id string) string {
	return metadataPrefix + id + ".json"
}

func normalizeFileIDs(fileIDs []string) []string {
	seen := make(map[string]struct{}, len(fileIDs))
	result := make([]string, 0, len(fileIDs))

	for _, fileID := range fileIDs {
		fileID = strings.TrimSpace(fileID)
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

func removeFileID(fileIDs []string, fileID string) []string {
	result := fileIDs[:0]
	for _, existingFileID := range fileIDs {
		if existingFileID != fileID {
			result = append(result, existingFileID)
		}
	}
	return result
}
