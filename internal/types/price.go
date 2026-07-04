package types

type Price int64

func (p Price) Valid() bool {
	return p > 0
}