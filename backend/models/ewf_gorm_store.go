package models

import (
	"context"

	"github.com/xmonader/ewf"
	"gorm.io/gorm"
)

type EWFGormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *EWFGormStore {
	return &EWFGormStore{db: db}
}

func (s *EWFGormStore) Setup() error {
	return s.db.AutoMigrate(&ewf.Workflow{})
}

func (s *EWFGormStore) SaveWorkflow(ctx context.Context, workflow *ewf.Workflow) error {
	return s.db.WithContext(ctx).Save(workflow).Error
}

func (s *EWFGormStore) LoadWorkflowByName(ctx context.Context, name string) (*ewf.Workflow, error) {
	var workflow ewf.Workflow
	if err := s.db.WithContext(ctx).Where("name = ?", name).First(&workflow).Error; err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (s *EWFGormStore) LoadWorkflowByUUID(ctx context.Context, uuid string) (*ewf.Workflow, error) {
	var workflow ewf.Workflow
	if err := s.db.WithContext(ctx).Where("uuid = ?", uuid).First(&workflow).Error; err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (s *EWFGormStore) ListWorkflowUUIDsByStatus(ctx context.Context, status ewf.WorkflowStatus) ([]string, error) {
	var uuids []string
	err := s.db.WithContext(ctx).
		Model(&ewf.Workflow{}).
		Where("status = ?", status).
		Pluck("uuid", &uuids).
		Error
	return uuids, err
}

func (s *EWFGormStore) Close() error {
	return nil
}
