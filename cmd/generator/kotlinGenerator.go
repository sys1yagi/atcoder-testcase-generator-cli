package generator

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

type kotlinGenerator struct {
	config Config
}

func (g kotlinGenerator) Generate(contestName string, problems []*Problem) error {
	contestName = strings.ToLower(contestName)
	for _, problem := range problems {
		// gen program.kt
		err := g.kotlinMain(contestName, problem)
		if err != nil {
			return err
		}

		// gen test.kt
		err = g.kotlinTest(contestName, problem)
		if err != nil {
			return err
		}

		// gen test files
		err = g.kotlinTestFile(contestName, problem)
		if err != nil {
			return err
		}
	}
	err := g.kotlinTestTool()
	if err != nil {
		return err
	}
	return nil
}

func (g kotlinGenerator) kotlinTestTool() error {
	test := template.Must(template.ParseFiles("template/kotlin/testtool.txt"))
	dir := fmt.Sprintf("%s/test/kotlin/testtool",
		g.config.DestinationDirPath,
	)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}
	testFile, err := os.Create(
		fmt.Sprintf("%s/StandardInOutRule.kt",
			dir,
		))
	if err != nil {
		return err
	}
	defer testFile.Close()

	if err := test.Execute(testFile, map[string]string{}); err != nil {
		log.Fatal(err)
	}
	return nil
}

type TestData struct {
	Package   string
	Problem   string
	InOutList string
}

func (g kotlinGenerator) kotlinTest(contestName string, problem *Problem) error {
	programName := strings.ToLower(problem.Name)
	dir := fmt.Sprintf("%s/test/kotlin/%s/%s/%s",
		g.config.DestinationDirPath,
		strings.ReplaceAll(g.config.KotlinPackageName, ".", "/"),
		contestName,
		programName,
	)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}
	testFile, err := os.Create(
		fmt.Sprintf("%s/%sKtTest.kt",
			dir,
			programName,
		))
	if err != nil {
		return err
	}
	defer testFile.Close()

	if len(problem.TestInputAndOutputs) == 0 {
		test := template.Must(template.ParseFiles("template/kotlin/test_default.txt"))
		data := map[string]TestData{
			"testData": {
				Package:   fmt.Sprintf("%s.%s.%s", g.config.KotlinPackageName, contestName, programName),
				Problem:   programName,
				InOutList: "",
			},
		}
		if err := test.Execute(testFile, data); err != nil {
			log.Fatal(err)
		}
	} else {
		test := template.Must(template.ParseFiles("template/kotlin/test.txt"))
		var inOutFileNames []string
		for _, io := range problem.TestInputAndOutputs {
			inOutFileNames = append(inOutFileNames, fmt.Sprintf("Arguments.arguments(\"%s\", \"%s\")", io.CaseName, io.CaseName))
		}
		inOutFileName := strings.Join(inOutFileNames, ",\n")

		data := map[string]TestData{
			"testData": {
				Package:   fmt.Sprintf("%s.%s.%s", g.config.KotlinPackageName, contestName, programName),
				Problem:   programName,
				InOutList: inOutFileName,
			},
		}

		if err := test.Execute(testFile, data); err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func (g kotlinGenerator) kotlinTestFile(contestName string, problem *Problem) error {
	for _, testCase := range problem.TestInputAndOutputs {
		programName := strings.ToLower(problem.Name)
		dir := fmt.Sprintf("%s/test/resources/%s/%s/%s",
			g.config.DestinationDirPath,
			strings.ReplaceAll(g.config.KotlinPackageName, ".", "/"),
			contestName,
			programName,
		)
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			return err
		}

		inFile, err := os.Create(
			fmt.Sprintf("%s/in_%s",
				dir,
				testCase.CaseName,
			))
		if err != nil {
			return err
		}
		defer inFile.Close()
		inFile.WriteString(testCase.Input)

		outFile, err := os.Create(
			fmt.Sprintf("%s/out_%s",
				dir,
				testCase.CaseName,
			))
		if err != nil {
			return err
		}
		defer outFile.Close()
		outFile.WriteString(testCase.Output)
	}
	return nil
}

func (g kotlinGenerator) kotlinMain(contestName string, problem *Problem) error {
	main := template.Must(template.ParseFiles("template/kotlin/main.txt"))
	programName := strings.ToLower(problem.Name)
	data := map[string]string{
		"package":     fmt.Sprintf("%s.%s.%s", g.config.KotlinPackageName, contestName, programName),
		"contestName": contestName,
		"problem":     programName,
	}

	dir := fmt.Sprintf("%s/main/kotlin/%s/%s/%s",
		g.config.DestinationDirPath,
		strings.ReplaceAll(g.config.KotlinPackageName, ".", "/"),
		contestName,
		programName,
	)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}
	mainFile, err := os.Create(
		fmt.Sprintf("%s/%s.kt",
			dir,
			programName,
		))
	if err != nil {
		return err
	}
	defer mainFile.Close()

	if err := main.Execute(mainFile, data); err != nil {
		log.Fatal(err)
	}
	return nil
}
