package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

var iddeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "remove identity tokens/keys",
	Long:  getIDAddLngUsage(),
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func deletePublicKeyGitlab(idInfo tIDInfo) {
	installID := mainViper.GetString("installation-id")
	glTemporaryToken := getTemporaryGitlabToken(idInfo)
	pubKey := getSSHPubKey()
	endPoint := "https://gitlab.com/api/v4/user/keys"
	body := strings.NewReader(fmt.Sprintf(`{
				"title": "%v",
				"key": "%v"
			}`, installID, strings.TrimRight(pubKey, "\n")))
	req, err := http.NewRequest("POST", endPoint, body)
	if err != nil {
		log.Fatal(err)
	}

	// req.SetBasicAuth(idInfo.Username, idInfo.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+glTemporaryToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		if resp.StatusCode == 400 {
			// just saying it's already been taken, probably
			return
		}
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("resp.StatusCode = %+v\n", resp.StatusCode)
		log.Fatal(string(b))
	}
}
