package domain

// AggregateVersion is a validated optimistic-concurrency token.
type AggregateVersion uint64

func NewAggregateVersion(value uint64) (AggregateVersion, error) {
	if value == 0 {
		return 0, ErrInvalidArgument
	}
	return AggregateVersion(value), nil
}

func (v AggregateVersion) Uint64() uint64 { return uint64(v) }
