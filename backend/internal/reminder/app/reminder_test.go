package app

import (
	"context"
	"testing"
	"time"

	model "github.com/docvault/backend/internal/reminder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubReminderRepository struct {
	getByDocumentReminders []model.ReminderRule
	listByTenantReminders  []model.ReminderRule
	listUpcomingReminders  []model.ReminderRule
	getByDocumentTenantID  string
	getByDocumentID        string
	listByTenantTenantID   string
	listByTenantActiveOnly bool
	listUpcomingTenantID   string
	listUpcomingWithinDays int
	getByDocumentCalled    bool
	listByTenantCalled     bool
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

func (s *stubReminderRepository) ListByTenant(_ context.Context, tenantID string, activeOnly bool) ([]model.ReminderRule, error) {
	s.listByTenantCalled = true
	s.listByTenantTenantID = tenantID
	s.listByTenantActiveOnly = activeOnly
	return s.listByTenantReminders, nil
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

func TestReminderServiceListUsesTenantListing(t *testing.T) {
	repo := &stubReminderRepository{
		listByTenantReminders: []model.ReminderRule{{ID: "rule-1"}},
	}

	svc := NewReminderService(repo)
	output, err := svc.List(context.Background(), &ListRemindersInput{
		TenantID: "tenant-1",
	})

	require.NoError(t, err)
	assert.True(t, repo.listByTenantCalled)
	assert.False(t, repo.getByDocumentCalled)
	assert.False(t, repo.listUpcomingCalled)
	assert.Equal(t, "tenant-1", repo.listByTenantTenantID)
	assert.False(t, repo.listByTenantActiveOnly)
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

func TestReminderServiceListActiveOnlyFiltersDocumentResults(t *testing.T) {
	repo := &stubReminderRepository{
		getByDocumentReminders: []model.ReminderRule{{ID: "rule-1", Active: true}, {ID: "rule-2", Active: false}},
	}

	svc := NewReminderService(repo)
	output, err := svc.List(context.Background(), &ListRemindersInput{
		TenantID:   "tenant-1",
		DocumentID: "doc-1",
		ActiveOnly: true,
	})

	require.NoError(t, err)
	assert.Len(t, output.Reminders, 1)
	assert.Equal(t, "rule-1", output.Reminders[0].ID)
}

func TestReminderServiceListActiveOnlyPassesFilterToTenantListing(t *testing.T) {
	repo := &stubReminderRepository{}

	svc := NewReminderService(repo)
	_, err := svc.List(context.Background(), &ListRemindersInput{
		TenantID:   "tenant-1",
		ActiveOnly: true,
	})

	require.NoError(t, err)
	assert.True(t, repo.listByTenantCalled)
	assert.True(t, repo.listByTenantActiveOnly)
}
