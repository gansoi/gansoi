package template

type (
	// Renderer must be implemented by types that would like to self-render.
	Renderer interface {
		RenderFunc(func() error)
	}
)
