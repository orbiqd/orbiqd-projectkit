package standard

type Repository interface {
	GetAll() ([]Standard, error)
	AddStandard(standard Standard) error
	RemoveAll() error
}
