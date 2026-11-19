package order

// State is intentionally small: each concrete state owns its allowed events.
type State interface {
	Status() Status
	Apply(Event) (State, error)
}
type orderState struct {
	status      Status
	transitions map[Event]Status
}

func (s orderState) Status() Status { return s.status }
func (s orderState) Apply(event Event) (State, error) {
	next, ok := s.transitions[event]
	if !ok {
		return nil, ErrInvalidTransition
	}
	return stateFor(next), nil
}
func stateFor(status Status) State {
	switch status {
	case StatusCreated:
		return orderState{status: StatusCreated, transitions: map[Event]Status{EventAssemble: StatusAssembling, EventCancel: StatusCancelled}}
	case StatusAssembling:
		return orderState{status: StatusAssembling, transitions: map[Event]Status{EventShip: StatusShipped, EventCancel: StatusCancelled}}
	case StatusShipped:
		return orderState{status: StatusShipped, transitions: map[Event]Status{EventDeliver: StatusDelivered}}
	case StatusDelivered:
		return orderState{status: StatusDelivered, transitions: map[Event]Status{}}
	case StatusCancelled:
		return orderState{status: StatusCancelled, transitions: map[Event]Status{}}
	default:
		return nil
	}
}
func NewState(status Status) (State, error) {
	s := stateFor(status)
	if s == nil {
		return nil, ErrInvalidTransition
	}
	return s, nil
}
func NextStatus(current Status, event Event) (Status, error) {
	state, err := NewState(current)
	if err != nil {
		return "", err
	}
	next, err := state.Apply(event)
	if err != nil {
		return "", err
	}
	return next.Status(), nil
}
