package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "walks you through your setup",
	Long:  `prompts for accounts you want to add, which package you wnat to use (as well as provides defaults)`,
	Run: func(cmd *cobra.Command, args []string) {
		if readConfirm("Add an id (github, gitlab, etc)? [y/N] ") {
			idsInitialize(initAddIDs())
		}

		if readConfirm("Install a zetup package? [y/N] ") {
			initAddZetupPkg()
		}
	},
}

func initAddZetupPkg() {
	fmt.Println("Note: Package strings must be in the form [hostname]/[username]/[reponame]\nFor Example: github.com/zetup-sh/zetup-pkg")
	userInput := readInput("Zetup pkg: ")

	usePkg(userInput)
}

var initIDsAlreadyAdded []string

func initAddIDs() []tIDInfo {
	var idsInfo []tIDInfo
	newIDInfo := getEmptyID()
	quickID := newIDInfo.Type + "/" + newIDInfo.Username
	if contains(initIDsAlreadyAdded, quickID) {
		fmt.Println("You can only have one active account of each type at a time.")
		fmt.Println("But you can easily switch between accounts later on with `zetup use`")
	} else {
		idsInfo = append(idsInfo, newIDInfo)
	}
	initIDsAlreadyAdded = append(initIDsAlreadyAdded, quickID)

	if readConfirm("Create another id? [y/N] ") {
		idsInfo = append(idsInfo, initAddIDs()...)
	}
	return idsInfo
}

func init() {
	rootCmd.AddCommand(initCmd)
}
