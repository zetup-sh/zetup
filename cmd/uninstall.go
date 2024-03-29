package cmd

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/bgentry/speakeasy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// initCmd represents the init command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "deletes github and local config",
	Long: `
	Remove the ssh key and personal access token from github and deletes config file
	`,
	Run: func(cmd *cobra.Command, args []string) {
		deleteSSHKey()
		deleteGithubToken()
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func deleteSSHKey() {
	sshKeyId := viper.GetString("ssh-key-id")
	if sshKeyId == "" {
		return
	}
	req, err := http.NewRequest("DELETE", "https://api.github.com/user/keys/"+sshKeyId, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(viper.GetString("github-username"), githubToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
		log.Fatal(string(b))
	}

	defer resp.Body.Close()
	viper.Set("ssh-key-id", nil)
	viper.WriteConfig()
}

func deleteGithubToken() {
	githubTokenId := viper.GetString("github-token-id")
	if githubTokenId == "" {
		return
	}
	password := viper.GetString("github-password")
	if password == "" {
		log.Println("Sorry, I can only delete the personal access token with your password.")
		password, err = speakeasy.Ask("Github Password: ")
		if err != nil {
			log.Fatal(err)
		}
	}
	req, err := http.NewRequest("DELETE", "https://api.github.com/authorizations/"+githubTokenId, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(viper.GetString("github-username"), password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
		log.Fatal(string(b))
	}

	defer resp.Body.Close()

	viper.Set("github-token-id", nil)
	viper.WriteConfig()

	err = os.Remove(viper.ConfigFileUsed())
	if err != nil {
		log.Fatal(err)
	}
}
