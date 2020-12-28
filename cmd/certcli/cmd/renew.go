package cmd

import (
	"log"
	"net/url"
	"time"

	"github.com/KalleDK/go-certapi/certapi"
	"github.com/KalleDK/go-certapi/cmd/certcli/certcli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func renewDomain(domain string, force bool, store certcli.DomainStore) error {
	info, cstore, err := store.Get(domain)
	if err != nil {
		return err
	}
	var state certapi.CertInfo
	if err := cstore.LoadState(&state); err != nil {
		return err
	}
	if !force && state.NextRenewTime.After(time.Now()) {
		log.Println(domain + " not renewtime yet " + state.NextRenewTime.Local().String())
		return nil
	}
	baseurl, err := url.Parse(info.Server)
	if err != nil {
		return err
	}
	newstate, err := fetchState(baseurl, domain)
	if err != nil {
		return err
	}
	if !force && newstate.Serial == state.Serial {
		log.Println(domain + " same serial")
		return nil
	}

	fullchain, err := fetchFullchain(baseurl, domain)
	if err != nil {
		return err
	}

	cert, err := fetchCertificate(baseurl, domain)
	if err != nil {
		return err
	}

	if err := cstore.SaveCertificate(cert); err != nil {
		return err
	}
	if err := cstore.SaveFullchain(fullchain); err != nil {
		return err
	}
	if err := cstore.SaveState(newstate); err != nil {
		return err
	}
	log.Println(domain + " renewed")
	return nil
}

// renewCmd represents the renew command
var renewCmd = &cobra.Command{
	Use:   "renew <domain>",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			log.Fatal(err)
		}

		store := certcli.DomainStore{Path: viper.GetString("dir")}
		if err := renewDomain(domain, force, store); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	renewCmd.Flags().BoolP("force", "f", false, "Force")
	rootCmd.AddCommand(renewCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// renewCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// renewCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
