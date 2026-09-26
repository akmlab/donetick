package circle

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// PointRules are the JSON-backed scoring weights for a circle leaderboard.
type PointRules struct {
	Completion          float64            `json:"completion"`
	Subtask             float64            `json:"subtask"`
	Skip                float64            `json:"skip"`
	PriorityMultipliers map[string]float64 `json:"priority_multipliers"`
}

// DefaultPointRules matches the product defaults for completion scoring.
func DefaultPointRules() PointRules {
	return PointRules{
		Completion: 1,
		Subtask:    1,
		Skip:       -1,
		PriorityMultipliers: map[string]float64{
			"0": 1,
			"1": 1,
			"2": 1.5,
			"3": 2,
			"4": 3,
		},
	}
}

// ParsePointRules parses circle.point_rules JSON. Empty string returns defaults.
// Missing priority multipliers are filled from the defaults.
func ParsePointRules(raw string) (PointRules, error) {
	defaults := DefaultPointRules()
	if raw == "" {
		return defaults, nil
	}
	var rules PointRules
	if err := json.Unmarshal([]byte(raw), &rules); err != nil {
		return PointRules{}, err
	}
	if rules.PriorityMultipliers == nil {
		rules.PriorityMultipliers = map[string]float64{}
	}
	for key, value := range defaults.PriorityMultipliers {
		if _, ok := rules.PriorityMultipliers[key]; !ok {
			rules.PriorityMultipliers[key] = value
		}
	}
	return rules, nil
}

// MarshalPointRules returns canonical JSON for persistence.
func MarshalPointRules(rules PointRules) (string, error) {
	if rules.PriorityMultipliers == nil {
		rules.PriorityMultipliers = DefaultPointRules().PriorityMultipliers
	}
	b, err := json.Marshal(rules)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// HistoryEvent is one scored completion or skip for the leaderboard.
type HistoryEvent struct {
	UserID   int
	Status   int // 1 complete, 2 skip
	Priority int
	Subtasks int
}

func multiplierFor(rules PointRules, priority int) float64 {
	key := strconv.Itoa(priority)
	if m, ok := rules.PriorityMultipliers[key]; ok {
		return m
	}
	return 1
}

// ScoreHistory returns total points per user id from history events.
func ScoreHistory(rules PointRules, events []HistoryEvent) map[int]float64 {
	scores := map[int]float64{}
	for _, event := range events {
		mult := multiplierFor(rules, event.Priority)
		switch event.Status {
		case 1:
			scores[event.UserID] += (rules.Completion + float64(event.Subtasks)*rules.Subtask) * mult
		case 2:
			scores[event.UserID] += rules.Skip * mult
		}
	}
	return scores
}

// FormatPoints is a helper for debugging/tests.
func FormatPoints(v float64) string {
	return fmt.Sprintf("%g", v)
}
