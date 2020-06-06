package generator

type TestFile struct {
	CaseName string
	Data     string
}

type TestInputAndOutput struct {
	CaseName string
	Input    string
	Output   string
}

type Problem struct {
	Name                string
	TestInputAndOutputs []TestInputAndOutput
}
