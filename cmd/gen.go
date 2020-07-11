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
	"github.com/sys1yagi/atcoder-testcase-generator-cli/cmd/generator"
	"io"
	"os"
	"strings"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "get in/out data of a contest then generate test cases",
	Long: `get in/out data of a contest then generate test cases`,
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

		// テストケースが存在しない場合
		if len(problems) == 0 {
			problems = []*generator.Problem{
				{
					Name: "A",
					TestInputAndOutputs: []generator.TestInputAndOutput{},
				},
				{
					Name: "B",
					TestInputAndOutputs: []generator.TestInputAndOutput{},
				},
				{
					Name: "C",
					TestInputAndOutputs: []generator.TestInputAndOutput{},
				},
				{
					Name: "D",
					TestInputAndOutputs: []generator.TestInputAndOutput{},
				},
				{
					Name: "E",
					TestInputAndOutputs: []generator.TestInputAndOutput{},
				},
				{
					Name: "F",
					TestInputAndOutputs: []generator.TestInputAndOutput{},
				},
			}

		}

		// ファイル群を生成する
		g := generator.GetGenerator("kotlin", config)
		if g == nil {
			return fmt.Errorf("generator not found")
		}
		return g.Generate(contestName, problems)
	},
}

var url = "https://www.dropbox.com/sh/nx3tnilzqz7df8a/AAAYlTq2tiEHl5hsESw6-yfLa?dl=0"

var sharedLink = &files.SharedLink{
	url,
	"",
}

// 指定したコンテスト名のテストの入出力データを取り出す
func getTestCases(contestName string) ([]*generator.Problem, error) {
	config := dropbox.Config{
		Token:    config.DropboxAccessToken,
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
			// フォルダ一覧からcontestNameに一致するものを探す
			if strings.ToLower(v.Name) == strings.ToLower(contestName) {
				return getTestCasesFromContestFolder(f, s, v)
			}
		default:
			// no op
		}
	}
	return nil, nil
}

// コンテストフォルダからテストデータを取り出す
func getTestCasesFromContestFolder(client files.Client, sharingClient sharing.Client, folder *files.FolderMetadata) (problems []*generator.Problem, err error) {
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

func getTestCasesFromProblemFolder(client files.Client, sharingClient sharing.Client, parentFolderName string, folder *files.FolderMetadata) (*generator.Problem, error) {
	path := fmt.Sprintf("/%s/%s", parentFolderName, folder.Name)
	args := files.NewListFolderArg(path)
	args.SharedLink = sharedLink
	res, err := client.ListFolder(args)
	if err != nil {
		return nil, err
	}

	problem := generator.Problem{
		Name: folder.Name,
	}

	var input []*generator.TestFile
	var output []*generator.TestFile

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
		var testInputAndOutput generator.TestInputAndOutput
		testInputAndOutput.CaseName = i.CaseName
		testInputAndOutput.Input = i.Data
		inName := strings.ReplaceAll(i.CaseName, ".in", "")
		for _, o := range output {
			outName := strings.ReplaceAll(o.CaseName, ".out", "")
			if inName == outName {
				testInputAndOutput.Output = o.Data
				problem.TestInputAndOutputs = append(problem.TestInputAndOutputs, testInputAndOutput)
				break
			}
		}
	}

	return &problem, nil
}

func getTestCasesFromProblemCaseFolder(client files.Client, sharingClient sharing.Client, parentFolderName string, folder *files.FolderMetadata) (testFiles []*generator.TestFile, err error) {
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
			testFile := generator.TestFile{}
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
