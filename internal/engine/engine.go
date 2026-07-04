package engine

type Engine interface{}

type MatchingEngine struct{}

func New() Engine {
	return &MatchingEngine{}
}
