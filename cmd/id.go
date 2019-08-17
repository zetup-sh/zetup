package cmd

import (
	"github.com/spf13/cobra"
)

// idCmd represents the id command
var idCmd = &cobra.Command{
	Use:   "id",
	Short: "Manage identities for zetup",
	Long:  `adds github identities`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("id called")
	// },
}

var addIDCmd = &cobra.Command{
	Use:   "add",
	Short: "add an identity",
	// Long: ``,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var idType string

var possibleIDTypes = []string{
	"github",
	"gitlab",
	"digitalocean",
}

func init() {
	rootCmd.AddCommand(idCmd)
	idCmd.AddCommand(addIDCmd)
	idCmd.PersistentFlags().StringVarP(&idType, "id-type", "t", "github", "id type")
}
