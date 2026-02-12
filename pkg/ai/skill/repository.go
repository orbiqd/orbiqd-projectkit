package skill

import "errors"

type Repository interface {
	GetSkillByName(name Name) (*Skill, error)
	GetAll() ([]Skill, error)
	AddSkill(skill Skill) error
	RemoveAll() error
}

var (
	ErrSkillAlreadyExists = errors.New("skill already exists")
	ErrSkillNotFound      = errors.New("skill not found")
)
