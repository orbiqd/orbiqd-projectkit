package instruction

// Repository provides access to stored AI instructions.
type Repository interface {
	// GetAll returns all stored instruction sets.
	GetAll() ([]Instructions, error)

	// AddInstructions stores the provided instruction set.
	AddInstructions(instructions Instructions) error
}
