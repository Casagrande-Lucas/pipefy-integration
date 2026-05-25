package valueobject

// Priority represents the processing priority of a client based on their patrimony value.
type Priority string

const (
	// PriorityHigh is assigned when patrimony value is >= 200,000.
	PriorityHigh Priority = "prioridade_alta"

	// PriorityNormal is assigned when patrimony value is < 200,000.
	PriorityNormal Priority = "prioridade_normal"

	// priorityThreshold is the minimum patrimony value for high priority.
	priorityThreshold = 200_000.0
)

// NewPriority calculates the priority based on the patrimony value.
// Business rule: value >= 200,000 → PriorityHigh; value < 200,000 → PriorityNormal.
func NewPriority(patrimonyValue float64) Priority {
	if patrimonyValue >= priorityThreshold {
		return PriorityHigh
	}
	return PriorityNormal
}

// String returns the priority as a plain string.
func (p Priority) String() string {
	return string(p)
}
