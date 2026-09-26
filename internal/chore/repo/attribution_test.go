package chore

import (
	"context"
	"fmt"
	"testing"
	"time"

	"donetick.com/core/config"
	chModel "donetick.com/core/internal/chore/model"
	cModel "donetick.com/core/internal/circle/model"
	stModel "donetick.com/core/internal/subtask/model"
	syncModel "donetick.com/core/internal/sync/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newAttributionTestRepository(t *testing.T) (*ChoreRepository, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&chModel.Chore{},
		&chModel.ChoreHistory{},
		&chModel.ChoreAssignees{},
		&stModel.SubTask{},
		&cModel.UserCircle{},
		&syncModel.SyncCursor{},
	))
	cfg := &config.Config{}
	cfg.Database.Type = "sqlite"
	return NewChoreRepository(db, cfg), db
}

func intPtr(i int) *int {
	return &i
}

func TestCompleteChoreAppliesHistoryCredit(t *testing.T) {
	repo, db := newAttributionTestRepository(t)
	ctx := context.Background()
	now := time.Now().UTC()
	chore := &chModel.Chore{ID: 1, CircleID: 1, Priority: 3, AssignedTo: intPtr(20), IsActive: true}
	require.NoError(t, db.Create(chore).Error)
	require.NoError(t, db.Create(&cModel.UserCircle{UserID: 20, CircleID: 1, IsActive: true}).Error)
	completedAt := now
	require.NoError(t, db.Create(&stModel.SubTask{
		ChoreID:     1,
		Name:        "step",
		CompletedAt: &completedAt,
		CompletedBy: 20,
	}).Error)

	due := now.Add(24 * time.Hour)
	credit := &chModel.HistoryAttribution{ActorUserID: 10, PerformedByUserID: 20}
	require.NoError(t, repo.CompleteChore(ctx, chore, nil, 20, &due, &now, intPtr(20), false, credit))

	var history chModel.ChoreHistory
	require.NoError(t, db.Where("chore_id = ?", 1).First(&history).Error)
	require.Equal(t, 20, history.CompletedBy)
	require.Equal(t, 10, history.ActorUserID)
	require.Equal(t, 20, history.PerformedByUserID)
	require.Equal(t, 3, history.PriorityAtCompletion)
	require.Equal(t, 1, history.SubtasksCompleted)
	require.Equal(t, chModel.ChoreHistoryStatusCompleted, history.Status)
}

func TestCompleteChoreNilCreditUsesUserID(t *testing.T) {
	repo, db := newAttributionTestRepository(t)
	ctx := context.Background()
	now := time.Now().UTC()
	chore := &chModel.Chore{ID: 2, CircleID: 1, Priority: 1, AssignedTo: intPtr(7), IsActive: true}
	require.NoError(t, db.Create(chore).Error)
	require.NoError(t, db.Create(&cModel.UserCircle{UserID: 7, CircleID: 1, IsActive: true}).Error)

	due := now.Add(24 * time.Hour)
	require.NoError(t, repo.CompleteChore(ctx, chore, nil, 7, &due, &now, intPtr(7), false, nil))

	var history chModel.ChoreHistory
	require.NoError(t, db.Where("chore_id = ?", 2).First(&history).Error)
	require.Equal(t, 7, history.CompletedBy)
	require.Equal(t, 7, history.ActorUserID)
	require.Equal(t, 7, history.PerformedByUserID)
	require.Equal(t, 1, history.PriorityAtCompletion)
}

func TestSkipChoreAppliesHistoryCredit(t *testing.T) {
	repo, db := newAttributionTestRepository(t)
	ctx := context.Background()
	now := time.Now().UTC()
	chore := &chModel.Chore{ID: 3, CircleID: 1, Priority: 4, AssignedTo: intPtr(20), IsActive: true}
	require.NoError(t, db.Create(chore).Error)

	due := now.Add(24 * time.Hour)
	credit := &chModel.HistoryAttribution{ActorUserID: 10, PerformedByUserID: 20}
	require.NoError(t, repo.SkipChore(ctx, chore, 20, &due, intPtr(20), &now, credit))

	var history chModel.ChoreHistory
	require.NoError(t, db.Where("chore_id = ?", 3).First(&history).Error)
	require.Equal(t, 20, history.CompletedBy)
	require.Equal(t, 10, history.ActorUserID)
	require.Equal(t, 20, history.PerformedByUserID)
	require.Equal(t, 4, history.PriorityAtCompletion)
	require.Equal(t, chModel.ChoreHistoryStatusSkipped, history.Status)
}
