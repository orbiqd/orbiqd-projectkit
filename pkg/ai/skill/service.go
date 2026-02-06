package skill

import "context"

type Service interface {
	GetByName(ctx context.Context, name Name) (*Skill, error)
}
