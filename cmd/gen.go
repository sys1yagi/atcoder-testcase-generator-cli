/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"fmt"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/sharing"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Println("Error Usage: atgc $contestName")
			os.Exit(1)
		}
		var contestName = args[0]
		fmt.Println(contestName)

		// 対象のテストケースを取り出す
		problems, err := getTestCases(contestName)
		if err != nil {
			return err
		}
		fmt.Println("problems")

		for _, problem := range problems {
			for _, inout :=range	problem.TestInputAndOutputs {
				fmt.Println(inout.CaseName)
				fmt.Printf("in: %s\n", inout.Input)
				fmt.Printf("out: %s\n", inout.Output)
			}
		}

		// ファイル群を生成する

		return nil
	},
}

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

var url = "https://www.dropbox.com/sh/nx3tnilzqz7df8a/AAAYlTq2tiEHl5hsESw6-yfLa?dl=0"

var sharedLink = &files.SharedLink{
	url,
	"",
}

func getTestCases(contestName string) ([]*Problem, error) {
	config := dropbox.Config{
		Token:    config.AccessToken,
		LogLevel: dropbox.LogOff,
	}
	f := files.New(config)
	s := sharing.New(config)
	args := files.NewListFolderArg("")
	args.SharedLink = sharedLink
	res, err := f.ListFolder(args)
	if err != nil {
		return nil, err
	}
	for _, e := range res.Entries {
		switch v := e.(type) {
		case *files.FolderMetadata:
			if strings.ToLower(v.Name) == strings.ToLower(contestName) {
				return getTestCasesFromContestFolder(f, s, v)
			}
		default:
			// no op
		}
	}
	return nil, nil
}

func getTestCasesFromContestFolder(client files.Client, sharingClient sharing.Client, folder *files.FolderMetadata) (problems []*Problem, err error) {
	args := files.NewListFolderArg(fmt.Sprintf("/%s", folder.Name))
	args.SharedLink = sharedLink
	res, err := client.ListFolder(args)
	if err != nil {
		return nil, err
	}
	for _, e := range res.Entries {
		switch v := e.(type) {
		case *files.FolderMetadata:
			problem, err := getTestCasesFromProblemFolder(client, sharingClient, folder.Name, v)
			if err != nil {
				return nil, err
			}
			if problem != nil {
				problems = append(problems, problem)
			}
			fmt.Println(v.Name)
		default:
			// no op
		}
	}
	return
}

func getTestCasesFromProblemFolder(client files.Client, sharingClient sharing.Client, parentFolderName string, folder *files.FolderMetadata) (*Problem, error) {
	path := fmt.Sprintf("/%s/%s", parentFolderName, folder.Name)
	args := files.NewListFolderArg(path)
	args.SharedLink = sharedLink
	res, err := client.ListFolder(args)
	if err != nil {
		return nil, err
	}

	problem := Problem{}

	var input []*TestFile
	var output []*TestFile

	for _, e := range res.Entries {
		switch v := e.(type) {
		case *files.FolderMetadata:
			if v.Name == "in" {
				input, err = getTestCasesFromProblemCaseFolder(client, sharingClient, path, v)
				if err != nil {
					return nil, err
				}
			} else if v.Name == "out" {
				output, err = getTestCasesFromProblemCaseFolder(client, sharingClient, path, v)
				if err != nil {
					return nil, err
				}
			}
		default:
			// no op
		}
	}

	for _, i := range input {
		var testInputAndOutput TestInputAndOutput
		testInputAndOutput.CaseName = i.CaseName
		testInputAndOutput.Input = i.Data

		for _, o := range output {
			if i.CaseName == o.CaseName {
				testInputAndOutput.Output = o.Data
				problem.TestInputAndOutputs = append(problem.TestInputAndOutputs, testInputAndOutput)
				break
			}
		}
	}

	return &problem, nil
}

func getTestCasesFromProblemCaseFolder(client files.Client, sharingClient sharing.Client, parentFolderName string, folder *files.FolderMetadata) (testFiles []*TestFile, err error) {
	path := fmt.Sprintf("%s/%s", parentFolderName, folder.Name)
	args := files.NewListFolderArg(path)
	args.SharedLink = sharedLink
	res, err := client.ListFolder(args)
	if err != nil {
		return nil, err
	}

	for _, e := range res.Entries {
		switch v := e.(type) {
		case *files.FileMetadata:
			testFile := TestFile{}
			testFile.CaseName = v.Name
			filePath := fmt.Sprintf("%s/%s", path, v.Name)

			fileArgs := sharing.NewGetSharedLinkMetadataArg(url)
			fileArgs.Path = filePath

			fmt.Printf("%s loading...\n", filePath)

			_, content, err := sharingClient.GetSharedLinkFile(fileArgs)
			if err != nil {
				return nil, err
			}
			data := StreamToString(content)
			testFile.Data = data
			testFiles = append(testFiles, &testFile)
		default:
			// no op
		}
	}
	return
}

func StreamToString(stream io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.String()
}

func init() {
	rootCmd.AddCommand(genCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// genCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// genCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
