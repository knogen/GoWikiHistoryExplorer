/*
Copyright © 2023 ider <admin@knogen.com>

从提取好的 reference data 过滤去重数据,暂时搁置开发
*/
package cmd

import (
	"fmt"

	"GoWikiHistoryExplorer/internal/logic/loadFromReferenceFile"

	"github.com/spf13/cobra"
)

// lfrfCmd represents the lfrf command
var lfrfCmd = &cobra.Command{
	Use:   "lfrf",
	Short: "load from reference file, discard",
	Long:  `从已经提取的 reference 数据中提取数据`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("lfrf called")
		loadFromReferenceFile.Main()
	},
}

func init() {
	rootCmd.AddCommand(lfrfCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// lfrfCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lfrfCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
