package generator

type Generator interface {
	Generate(contestName string, problems []*Problem) error
}

func GetGenerator(name string, config Config) (g Generator) {
	switch name {
	case "kotlin":
		g = kotlinGenerator{
			config: config,
		}
	}
	return
}
