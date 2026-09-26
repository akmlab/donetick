package circle

import (
	"context"
	"fmt"
	"testing"

	cModel "donetick.com/core/internal/circle/model"
	cRepo "donetick.com/core/internal/circle/repo"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestScoreHistoryDefaultRules(t *testing.T) {
	rules := DefaultPointRules()
	events := []HistoryEvent{
		{UserID: 1, Status: 1, Priority: 4, Subtasks: 2},
		{UserID: 1, Status: 1, Priority: 2, Subtasks: 0},
		{UserID: 1, Status: 2, Priority: 3, Subtasks: 0},
	}
	scores := ScoreHistory(rules, events)
	if scores[1] != 8.5 {
		t.Fatalf("default rules: got %v, want 8.5", scores[1])
	}
}

func TestScoreHistoryCustomRules(t *testing.T) {
	rules := PointRules{
		Completion: 2,
		Subtask:    0,
		Skip:       0,
		PriorityMultipliers: map[string]float64{
			"0": 1, "1": 1, "2": 1, "3": 1, "4": 1,
		},
	}
	events := []HistoryEvent{
		{UserID: 1, Status: 1, Priority: 4, Subtasks: 2},
		{UserID: 1, Status: 1, Priority: 2, Subtasks: 0},
		{UserID: 1, Status: 2, Priority: 3, Subtasks: 0},
	}
	scores := ScoreHistory(rules, events)
	if scores[1] != 4 {
		t.Fatalf("custom rules: got %v, want 4", scores[1])
	}
}

func TestUpdatePointRulesRoundTrip(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&cModel.Circle{}))

	repo := cRepo.NewCircleRepository(db)
	circle, err := repo.CreateCircle(context.Background(), &cModel.Circle{Name: "test", CreatedBy: 1})
	require.NoError(t, err)

	rules := PointRules{
		Completion: 2,
		Subtask:    0,
		Skip:       0,
		PriorityMultipliers: map[string]float64{
			"0": 1, "1": 1, "2": 1, "3": 1, "4": 1,
		},
	}
	raw, err := MarshalPointRules(rules)
	require.NoError(t, err)
	require.NoError(t, repo.UpdatePointRules(context.Background(), circle.ID, raw))

	loaded, err := repo.GetCircleByID(context.Background(), circle.ID)
	require.NoError(t, err)
	parsed, err := ParsePointRules(loaded.PointRules)
	require.NoError(t, err)

	events := []HistoryEvent{
		{UserID: 1, Status: 1, Priority: 4, Subtasks: 2},
		{UserID: 1, Status: 1, Priority: 2, Subtasks: 0},
		{UserID: 1, Status: 2, Priority: 3, Subtasks: 0},
	}
	scores := ScoreHistory(parsed, events)
	if scores[1] != 4 {
		t.Fatalf("persisted rules: got %v, want 4", scores[1])
	}
}
