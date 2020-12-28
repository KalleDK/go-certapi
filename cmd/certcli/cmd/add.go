package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"

	"github.com/KalleDK/go-certapi/certapi"
	"github.com/KalleDK/go-certapi/cmd/certcli/certcli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
)

func fetchState(baseurl *url.URL, domain string) (state certapi.CertInfo, err error) {
	url, err := baseurl.Parse(path.Join("cert", domain))
	if err != nil {
		return
	}
	resp, err := http.Get(url.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(body, &state); err != nil {
		return
	}

	return state, nil
}

func fetchCertificate(baseurl *url.URL, domain string) (data []byte, err error) {
	url, err := baseurl.Parse(path.Join("cert", domain, "certificate"))
	if err != nil {
		return
	}
	resp, err := http.Get(url.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func fetchFullchain(baseurl *url.URL, domain string) (data []byte, err error) {
	url, err := baseurl.Parse(path.Join("cert", domain, "fullchain"))
	if err != nil {
		return
	}
	resp, err := http.Get(url.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func fetchCertificateKey(baseurl *url.URL, domain string, pass string) (data []byte, err error) {
	url, err := baseurl.Parse(path.Join("cert", domain, "key"))
	if err != nil {
		return
	}
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+certcli.MakeBearer(pass))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <domain>",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		domain := args[0]
		dir := viper.GetString("dir")
		rcmd, err := cmd.Flags().GetString("reloadcmd")
		if err != nil {
			log.Panic(err)
		}
		rargs, err := cmd.Flags().GetStringSlice("arg")
		if err != nil {
			log.Panic(err)
		}

		dstore := certcli.DomainStore{Path: dir}

		var pass string
		{
			fmt.Print("Enter passphrase: ")
			b, err := terminal.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				log.Panic(err)
			}
			fmt.Println()
			pass = string(b)
		}

		info := certcli.CertInfo{
			Server: viper.GetString("server"),
		}
		if rcmd != "" {
			info.ReloadCmd = rcmd
			info.Args = rargs
		}

		baseurl, err := url.Parse(info.Server)
		if err != nil {
			log.Panic(err)
		}

		key, err := fetchCertificateKey(baseurl, domain, pass)
		if err != nil {
			log.Panic(err)
		}

		state, err := fetchState(baseurl, domain)
		if err != nil {
			log.Panic(err)
		}

		fullchain, err := fetchFullchain(baseurl, domain)
		if err != nil {
			log.Panic(err)
		}

		cert, err := fetchCertificate(baseurl, domain)
		if err != nil {
			log.Panic(err)
		}

		{
			cstore, err := dstore.Add(domain, info)
			if err != nil {
				log.Panic(err)
			}

			if err := cstore.SaveKey(key); err != nil {
				log.Panic(err)
			}
			if err := cstore.SaveCertificate(cert); err != nil {
				log.Panic(err)
			}
			if err := cstore.SaveFullchain(fullchain); err != nil {
				log.Panic(err)
			}
			if err := cstore.SaveState(state); err != nil {
				log.Panic(err)
			}

			if info.ReloadCmd != "" {
				cmd := exec.Command(info.ReloadCmd, info.Args...)
				cmd.Env = append(os.Environ(), cstore.Env()...)
				stdoutStderr, err := cmd.CombinedOutput()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("%s\n", stdoutStderr)
			}
		}

	},
}

func init() {
	addCmd.Flags().StringSliceP("arg", "a", nil, "Args to reload command")
	addCmd.Flags().StringP("reloadcmd", "r", "", "ReloadCmd")
	addCmd.Flags().StringP("server", "s", "https://ca.example.com", "Certification Server")
	viper.BindPFlag("server", addCmd.Flags().Lookup("server"))
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
