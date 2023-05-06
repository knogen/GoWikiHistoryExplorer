/*
Copyright © 2023 ider <admin@knogen.com>

wikipedia reference try to match arxiv
*/
package cmd

import (
	"fmt"

	"GoWikiHistoryExplorer/internal/logic/matchArxiv"

	"github.com/spf13/cobra"
)

// mtavCmd represents the mtav command
var mtavCmd = &cobra.Command{
	Use:   "mtav",
	Short: "wikipedia reference try to match arxiv",
	Long:  `the source data is from history_ref_20220201_v2_arxiv, try to match arxiv and save result to elem match`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mtav called")
		matchArxiv.Main()
	},
}

func init() {
	rootCmd.AddCommand(mtavCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mtavCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mtavCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
