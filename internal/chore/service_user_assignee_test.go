package chore

import (
	"testing"

	chModel "donetick.com/core/internal/chore/model"
	circle "donetick.com/core/internal/circle/model"
)

func TestRotationAssigneesSkipsServiceUserWhenEmpty(t *testing.T) {
	humanA, humanB, service := 1, 2, 99
	members := []*circle.UserCircleDetail{
		{UserCircle: circle.UserCircle{UserID: humanA}},
		{UserCircle: circle.UserCircle{UserID: humanB}},
		{UserCircle: circle.UserCircle{UserID: service}, IsServiceUser: true},
	}

	for _, strategy := range []chModel.AssignmentStrategy{
		chModel.AssignmentStrategyRoundRobin,
		chModel.AssignmentStrategyLeastCompleted,
	} {
		chore := &chModel.Chore{
			ID:             1,
			AssignedTo:     intPtr(humanA),
			AssignStrategy: strategy,
			Assignees:      nil,
		}
		next, err := checkNextAssignee(chore, nil, humanA, members)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", strategy, err)
		}
		if next == nil {
			t.Fatalf("%s: expected next assignee", strategy)
		}
		if *next == service {
			t.Errorf("%s: service user must not be selected from empty assignee list", strategy)
		}
	}
}

func TestRotationAssigneesKeepsExplicitServiceUser(t *testing.T) {
	human, service := 1, 99
	members := []*circle.UserCircleDetail{
		{UserCircle: circle.UserCircle{UserID: human}},
		{UserCircle: circle.UserCircle{UserID: service}, IsServiceUser: true},
	}

	chore := &chModel.Chore{
		ID:             1,
		AssignedTo:     intPtr(human),
		AssignStrategy: chModel.AssignmentStrategyRoundRobin,
		Assignees: []chModel.ChoreAssignees{
			{ChoreID: 1, UserID: human},
			{ChoreID: 1, UserID: service},
		},
	}

	next, err := checkNextAssignee(chore, nil, human, members)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next == nil {
		t.Fatal("expected next assignee")
	}
	if *next != service {
		t.Errorf("expected round robin from human to service user %d, got %d", service, *next)
	}
}

func TestPerformerForAction(t *testing.T) {
	actor := 10
	member := 20
	inactive := 30
	members := []*circle.UserCircleDetail{
		{UserCircle: circle.UserCircle{UserID: member, IsActive: true}},
		{UserCircle: circle.UserCircle{UserID: inactive, IsActive: false}},
	}

	got, err := performerForAction(actor, nil, members)
	if err != nil || got != actor {
		t.Fatalf("omitted as_user_id: got %d err %v, want actor %d", got, err, actor)
	}

	zero := 0
	got, err = performerForAction(actor, &zero, members)
	if err != nil || got != actor {
		t.Fatalf("zero as_user_id: got %d err %v, want actor %d", got, err, actor)
	}

	got, err = performerForAction(actor, &member, members)
	if err != nil || got != member {
		t.Fatalf("active member: got %d err %v, want %d", got, err, member)
	}

	unknown := 999
	_, err = performerForAction(actor, &unknown, members)
	if err != errNotCircleMember {
		t.Fatalf("unknown id: got err %v, want errNotCircleMember", err)
	}

	_, err = performerForAction(actor, &inactive, members)
	if err != errNotCircleMember {
		t.Fatalf("inactive member: got err %v, want errNotCircleMember", err)
	}
}
