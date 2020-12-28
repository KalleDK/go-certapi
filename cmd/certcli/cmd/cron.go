package cmd

import (
	"log"

	"github.com/KalleDK/go-certapi/cmd/certcli/certcli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cronCmd represents the cron command
var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		store := certcli.DomainStore{Path: viper.GetString("dir")}
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			log.Fatal(err)
		}
		var config certcli.Config
		if err := store.LoadConfig(&config); err != nil {
			log.Fatal(err)
		}
		for domain := range config.Certs {
			renewDomain(domain, force, store)
		}
	},
}

func init() {
	cronCmd.Flags().BoolP("force", "f", false, "Force")
	rootCmd.AddCommand(cronCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cronCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cronCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
