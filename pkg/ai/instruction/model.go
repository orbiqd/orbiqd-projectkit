package instruction

type Rule string

type Category string

type Instructions struct {
	Category Category `json:"category" validate:"required"`
	Rules    []Rule   `json:"rules" validate:"required,min=1"`
}
