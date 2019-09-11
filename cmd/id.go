package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
)

// idCmd represents the id command
var idCmd = &cobra.Command{
	Use:   "id",
	Short: "Manage identities for zetup",
	// Long:  `adds  identities`,
	// Run: func(cmd *cobra.Command, args []string) {

	// },
}

func init() {
	idFile = filepath.Join(zetupDir, "identities.yml")
	rootCmd.AddCommand(idCmd)

	idCmd.AddCommand(idUseCmd)
	idUseCmd.Flags().BoolVarP(&idAddAddSSH, "ssh", "", true, "add ssh key to account")
	idUseCmd.Flags().BoolVarP(&idAddOverwrite, "overwrite", "", false, "overwrite existing accounts with the same username")
	idUseCmd.Flags().BoolVarP(&idAddGHToken, "gh-token", "", true, "create token for github instead of using plain text password")

	idCmd.AddCommand(idDeleteCmd)
}

// creates tokens, adds public keys, etc.
func idsInitialize(idsInfo []tIDInfo) {
	for _, idInfo := range idsInfo {
		if idInfo.Type == "github" {
			if !checkIsGithubToken(idInfo.Password) && idAddGHToken {
				tokenData := ensureGithubToken(idInfo)
				idInfo.Password = tokenData.Token
			}
			// authorization will fail here if they provided an incorrect password
			ensurePublicKeyGithub(idInfo)
		}
		if idInfo.Type == "gitlab" {
			ensurePublicKeyGitlab(idInfo)
		}

		mainViper.Set(idInfo.Type+".password", idInfo.Password)
		mainViper.Set(idInfo.Type+".username", idInfo.Username)

		// this will skip if it's already been set
		getUserInfoFromGitlab(idInfo)
		getUserInfoFromGithub()
		writeGitConfig()

		mainViper.WriteConfig()
	}
}

type ghAuthInfo struct {
	ID  int           `json:"id"`
	App ghAuthAppInfo `json:"app"`
}

type ghAuthAppInfo struct {
	Name string `json:"name"`
}

var githubAPIBase = "https://api.github.com"
var githubAPIAuthorizations = githubAPIBase + "/authorizations"

type idCredentials struct {
	Gitlab tIDInfo
	Github tIDInfo
}

type tIDInfo struct {
	Type     string `yaml:"type"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type tIDLists struct {
	List map[string][]tIDInfo
}
