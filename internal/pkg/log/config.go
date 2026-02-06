package log

type Config struct {
	Level  string `short:"v" help:"Log level" default:"info" enum:"debug,info,warn,error"`
	Format string `help:"Log format" default:"text-color"`
	Quiet  bool   `help:"Disable logging output"`
}
