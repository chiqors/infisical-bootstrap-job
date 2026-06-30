package bootstrap

func Run() error {
	cfg := LoadConfig()

	switch cfg.Mode {
	case ModePlatform:
		return runPlatform(cfg)
	case ModeOperator:
		return runOperator(cfg)
	case ModeProject:
		return runProject(cfg)
	default:
		panic("unsupported bootstrap mode")
	}
}
