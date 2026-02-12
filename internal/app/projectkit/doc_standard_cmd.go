package projectkit

type DocStandardCmd struct {
	Validate DocStandardValidateCmd `cmd:"validate" help:"Validate standard."`
	Render   DocStandardRenderCmd   `cmd:"render" help:"Render standards to documentation files."`
}
