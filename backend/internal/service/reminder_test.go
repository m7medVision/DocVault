package service

import (
	"context"
	"testing"
	"time"

	"github.com/docvault/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubReminderRepository struct {
	getByDocumentReminders []model.ReminderRule
	listUpcomingReminders  []model.ReminderRule
	getByDocumentTenantID  string
	getByDocumentID        string
	listUpcomingTenantID   string
	listUpcomingWithinDays int
	getByDocumentCalled    bool
	listUpcomingCalled     bool
}

func (s *stubReminderRepository) Create(context.Context, *model.ReminderRule) error {
	return nil
}

func (s *stubReminderRepository) GetByID(context.Context, string, string) (*model.ReminderRule, error) {
	return nil, nil
}

func (s *stubReminderRepository) GetByDocument(_ context.Context, tenantID, documentID string) ([]model.ReminderRule, error) {
	s.getByDocumentCalled = true
	s.getByDocumentTenantID = tenantID
	s.getByDocumentID = documentID
	return s.getByDocumentReminders, nil
}

func (s *stubReminderRepository) ListUpcoming(_ context.Context, tenantID string, withinDays int) ([]model.ReminderRule, error) {
	s.listUpcomingCalled = true
	s.listUpcomingTenantID = tenantID
	s.listUpcomingWithinDays = withinDays
	return s.listUpcomingReminders, nil
}

func (s *stubReminderRepository) Update(context.Context, *model.ReminderRule) error {
	return nil
}

func (s *stubReminderRepository) Delete(context.Context, string, string) error {
	return nil
}

func (s *stubReminderRepository) CreateEvent(context.Context, *model.ReminderEvent) error {
	return nil
}

func (s *stubReminderRepository) UpdateEvent(context.Context, *model.ReminderEvent) error {
	return nil
}

func (s *stubReminderRepository) GetPendingEvents(context.Context, time.Time) ([]model.ReminderEvent, error) {
	return nil, nil
}

func TestReminderServiceListUsesUpcomingWindow(t *testing.T) {
	repo := &stubReminderRepository{
		listUpcomingReminders: []model.ReminderRule{{ID: "rule-1"}},
	}

	svc := NewReminderService(repo)
	output, err := svc.List(context.Background(), &ListRemindersInput{
		TenantID: "tenant-1",
	})

	require.NoError(t, err)
	assert.True(t, repo.listUpcomingCalled)
	assert.False(t, repo.getByDocumentCalled)
	assert.Equal(t, "tenant-1", repo.listUpcomingTenantID)
	assert.Equal(t, 30, repo.listUpcomingWithinDays)
	assert.Len(t, output.Reminders, 1)
	assert.Equal(t, "rule-1", output.Reminders[0].ID)
}

func TestReminderServiceListByDocument(t *testing.T) {
	repo := &stubReminderRepository{
		getByDocumentReminders: []model.ReminderRule{{ID: "rule-2"}},
	}

	svc := NewReminderService(repo)
	output, err := svc.List(context.Background(), &ListRemindersInput{
		TenantID:   "tenant-1",
		DocumentID: "doc-1",
	})

	require.NoError(t, err)
	assert.True(t, repo.getByDocumentCalled)
	assert.False(t, repo.listUpcomingCalled)
	assert.Equal(t, "tenant-1", repo.getByDocumentTenantID)
	assert.Equal(t, "doc-1", repo.getByDocumentID)
	assert.Len(t, output.Reminders, 1)
	assert.Equal(t, "rule-2", output.Reminders[0].ID)
}
