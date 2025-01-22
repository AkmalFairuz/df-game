package game

type State uint8

const (
	StateUnknown State = iota
	StateWaiting
	StatePlaying
	StateFinished
)

func (state State) Waiting() bool {
	return state == StateWaiting
}

func (state State) Playing() bool {
	return state == StatePlaying
}

func (state State) Finished() bool {
	return state == StateFinished
}

func (state State) Unknown() bool {
	return state == StateUnknown
}

type ParticipantState uint8

const (
	ParticipantStateUnknown ParticipantState = iota
	ParticipantStatePlaying
	ParticipantStateSpectating
)

func (state ParticipantState) Playing() bool {
	return state == ParticipantStatePlaying
}

func (state ParticipantState) Spectating() bool {
	return state == ParticipantStateSpectating
}

func (state ParticipantState) Unknown() bool {
	return state == ParticipantStateUnknown
}
