package circuit

// ThresholdTripFunc .
func ThresholdTripFunc(threshold int64) TripFunc {
	return func(c Counter) bool {
		counts := c.Counts()
		return counts.Errors() >= threshold
	}
}

// ConsecutiveTripFunc .
func ConsecutiveTripFunc(threshold int64) TripFunc {
	return func(c Counter) bool {
		counts := c.Counts()
		return counts.ConseErrors >= threshold
	}
}

// RateTripFunc .
func RateTripFunc(rate float64, minSamples int64) TripFunc {
	return func(c Counter) bool {
		counts := c.Counts()
		return counts.Samples() >= minSamples && counts.ErrorRate() >= rate
	}
}
