package order

import "testing"

func TestStateMachineTransitions(t *testing.T) {
	tests := []struct {
		from  Status
		event Event
		want  Status
	}{
		{StatusCreated, EventAssemble, StatusAssembling},
		{StatusCreated, EventCancel, StatusCancelled},
		{StatusAssembling, EventShip, StatusShipped},
		{StatusAssembling, EventCancel, StatusCancelled},
		{StatusShipped, EventDeliver, StatusDelivered},
	}
	for _, tt := range tests {
		got, err := NextStatus(tt.from, tt.event)
		if err != nil {
			t.Fatalf("%s + %s: %v", tt.from, tt.event, err)
		}
		if got != tt.want {
			t.Fatalf("%s + %s: got %s want %s", tt.from, tt.event, got, tt.want)
		}
	}
}
func TestStateMachineRejectsInvalidTransitions(t *testing.T) {
	tests := []struct {
		from  Status
		event Event
	}{
		{StatusCreated, EventShip}, {StatusShipped, EventCancel}, {StatusDelivered, EventCancel}, {StatusCancelled, EventAssemble},
	}
	for _, tt := range tests {
		if _, err := NextStatus(tt.from, tt.event); err != ErrInvalidTransition {
			t.Fatalf("%s + %s: expected ErrInvalidTransition, got %v", tt.from, tt.event, err)
		}
	}
}
