package staging

type resourceContainer struct {
	// Essence is both a resource that is needed for most production-related tasks
	// and something that keeps the colony agents alive.
	// The essence can be collected externally by gathering or hunting.
	// If colony is out of essence, bad things may start to happen.
	Essence float64
}
