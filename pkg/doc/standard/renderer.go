package standard

type Renderer interface {
	Render(standard Standard) ([]byte, error)
	FileExtension() string
}
