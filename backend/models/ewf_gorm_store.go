package models

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/xmonader/ewf"
	"gorm.io/gorm"
)

type EWFGormStore struct {
	db *gorm.DB
}

type gormWorkflowRecord struct {
	UUID   string `gorm:"primaryKey;column:uuid"`
	Name   string `gorm:"column:name;not null;index"`
	Status string `gorm:"column:status;not null;index"`
	Data   []byte `gorm:"column:data;not null"`
}

func NewGormStore(db *gorm.DB) *EWFGormStore {
	return &EWFGormStore{db: db}
}

func (s *EWFGormStore) Setup() error {
	return s.db.AutoMigrate(&gormWorkflowRecord{})
}

func (s *EWFGormStore) SaveWorkflow(ctx context.Context, workflow *ewf.Workflow) error {
	data, err := json.Marshal(workflow)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow: %w", err)
	}

	gormWorkflow := gormWorkflowRecord{
		UUID:   workflow.UUID,
		Name:   workflow.Name,
		Status: string(workflow.Status),
		Data:   data,
	}

	return s.db.WithContext(ctx).Save(gormWorkflow).Error
}

func (s *EWFGormStore) LoadWorkflowByName(ctx context.Context, name string) (*ewf.Workflow, error) {
	var gormWorkflow gormWorkflowRecord
	if err := s.db.WithContext(ctx).Where("name = ?", name).First(&gormWorkflow).Error; err != nil {
		return nil, err
	}

	var workflow ewf.Workflow
	err := json.Unmarshal(gormWorkflow.Data, &workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow: %w", err)
	}
	return &workflow, nil
}

func (s *EWFGormStore) LoadWorkflowByUUID(ctx context.Context, uuid string) (*ewf.Workflow, error) {
	var gormWorkflow gormWorkflowRecord
	if err := s.db.WithContext(ctx).Where("uuid = ?", uuid).First(&gormWorkflow).Error; err != nil {
		return nil, err
	}
	var workflow ewf.Workflow
	if err := json.Unmarshal(gormWorkflow.Data, &workflow); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow: %w", err)
	}
	return &workflow, nil
}

func (s *EWFGormStore) ListWorkflowUUIDsByStatus(ctx context.Context, status ewf.WorkflowStatus) ([]string, error) {
	var uuids []string
	err := s.db.WithContext(ctx).
		Model(&gormWorkflowRecord{}).
		Where("status = ?", status).
		Pluck("uuid", &uuids).
		Error
	return uuids, err
}

func (s *EWFGormStore) Close() error {
	return nil
}
