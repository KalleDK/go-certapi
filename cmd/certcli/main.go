package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/KalleDK/go-certapi/certapi"
)

type Config struct {
	Server string
	Certs  map[string]interface{}
}

func reloadcmd() {
	fmt.Println(os.Args)
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", stdoutStderr)
}

func loadConfig(cpath string, config *Config) error {
	b, err := ioutil.ReadFile(cpath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, config); err != nil {
		return err
	}

	return nil
}

func saveConfig(cpath string, conf Config) error {
	fp, err := os.Create(cpath)
	if err != nil {
		return err
	}
	defer fp.Close()
	dec := json.NewEncoder(fp)
	dec.SetIndent("", "  ")
	return dec.Encode(conf)
}

func fetchState(dpath string, conf Config, domain string) error {
	statepath := filepath.Join(dpath, "state")
	var state certapi.CertInfo

	req, err := http.Get(conf.Server + "/cert/" + domain + "/")
	if err != nil {
		return err
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &state); err != nil {
		return err
	}

	fp, err := os.Create(statepath)
	if err != nil {
		return err
	}
	defer fp.Close()
	enc := json.NewEncoder(fp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(state); err != nil {
		return err
	}
	return nil
}

func fetchCertificate(dpath string, conf Config, domain string) error {
	req, err := http.Get(conf.Server + "/cert/" + domain + "/certificate")
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	certpath := filepath.Join(dpath, "server.cer")
	if err := ioutil.WriteFile(certpath, body, 755); err != nil {
		return err
	}
	return nil
}

func fetchFullchain(dpath string, conf Config, domain string) error {
	req, err := http.Get(conf.Server + "/cert/" + domain + "/fullchain")
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	certpath := filepath.Join(dpath, "fullchain.cer")
	if err := ioutil.WriteFile(certpath, body, 755); err != nil {
		return err
	}
	return nil
}

func fetchCertificateKey(dpath string, conf Config, domain string, pass string) error {
	sha := sha256.Sum256([]byte(pass))
	key := hex.EncodeToString(sha[:])

	requ, err := http.NewRequest("GET", conf.Server+"/cert/"+domain+"/key", nil)
	if err != nil {
		return err
	}
	requ.Header.Set("Authorization", "Bearer "+key)

	req, err := http.DefaultClient.Do(requ)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	certpath := filepath.Join(dpath, "server.key")
	if err := ioutil.WriteFile(certpath, body, 755); err != nil {
		return err
	}
	return nil
}

func renewDomain(spath string, domain string, force bool) (bool, error) {
	return false, nil
}

func createDomain(spath string, cpath string, domain string, pass string) {
	var conf Config
	if err := loadConfig(cpath, &conf); err != nil {
		log.Fatal(err)
	}

	_, ok := conf.Certs[domain]
	if ok {
		log.Fatal("domain already exists")
		return
	}

	dpath := filepath.Join(spath, domain)

	os.Mkdir(dpath, 755)

	if err := fetchState(dpath, conf, domain); err != nil {
		log.Fatal(err)
	}

	if err := fetchFullchain(dpath, conf, domain); err != nil {
		log.Fatal(err)
	}

	if err := fetchCertificate(dpath, conf, domain); err != nil {
		log.Fatal(err)
	}

	if err := fetchCertificateKey(dpath, conf, domain, pass); err != nil {
		log.Fatal(err)
	}

	conf.Certs[domain] = map[string]string{}

	if err := saveConfig(cpath, conf); err != nil {
		log.Fatal(err)
	}
}

func deleteDomain(spath string, cpath string, domain string) {
	var conf Config
	if err := loadConfig(cpath, &conf); err != nil {
		log.Fatal(err)
	}

	_, ok := conf.Certs[domain]
	if !ok {
		log.Print("domain doesn't exists")
		return
	}

	os.RemoveAll(filepath.Join(spath, domain))

	delete(conf.Certs, domain)

	if err := saveConfig(cpath, conf); err != nil {
		log.Fatal(err)
	}
}

func demo(cfile, spath string) {

	var conf Config
	if err := loadConfig(cfile, &conf); err != nil {
		log.Fatal(err)
	}

	for domain, _ := range conf.Certs {
		dpath := filepath.Join(spath, domain)
		if _, err := os.Stat(dpath); os.IsNotExist(err) {
			os.Mkdir(dpath, 755)
		}
		pull := false
		serial := ""
		statepath := filepath.Join(spath, domain, "state")
		if _, err := os.Stat(statepath); os.IsNotExist(err) {
			pull = true
			serial = ""
		} else {
			b, err := ioutil.ReadFile(statepath)
			if err != nil {
				log.Fatal(err)
			}
			var state certapi.CertInfo
			if err := json.Unmarshal(b, &state); err != nil {
				log.Fatal(err)
			}
			serial = state.Serial
			if state.NextRenewTime.Before(time.Now()) {
				pull = true
			}
		}
		if pull {
			var newstate certapi.CertInfo
			{
				req, err := http.Get(conf.Server + "/cert/" + domain + "/")
				if err != nil {
					log.Fatal(err)
				}
				defer req.Body.Close()
				body, err := ioutil.ReadAll(req.Body)
				if err != nil {
					log.Fatal(err)
				}
				if err := json.Unmarshal(body, &newstate); err != nil {
					log.Fatal(err)
				}
			}

			if newstate.Serial != serial {
				{
					b, err := json.Marshal(newstate)
					if err != nil {
						log.Fatal(err)
					}
					if err := ioutil.WriteFile(statepath, b, 755); err != nil {
						log.Fatal(err)
					}
				}
				{
					req, err := http.Get(conf.Server + "/cert/" + domain + "/fullchain")
					if err != nil {
						log.Fatal(err)
					}
					body, err := ioutil.ReadAll(req.Body)
					if err != nil {
						log.Fatal(err)
					}
					certpath := filepath.Join(spath, domain, "fullchain.cer")
					if err := ioutil.WriteFile(certpath, body, 755); err != nil {
						log.Fatal(err)
					}
				}
				{
					req, err := http.Get(conf.Server + "/cert/" + domain + "/certificate")
					if err != nil {
						log.Fatal(err)
					}
					body, err := ioutil.ReadAll(req.Body)
					if err != nil {
						log.Fatal(err)
					}
					certpath := filepath.Join(spath, domain, "server.cer")
					if err := ioutil.WriteFile(certpath, body, 755); err != nil {
						log.Fatal(err)
					}
				}
			}

		}
	}

}

func main() {
	//cfile := os.Args[2]
	cpath := filepath.Clean("certcli.conf")

	spath := filepath.Clean("certs")

	domain := "example.com"

	pass := "pass"

	deleteDomain(spath, cpath, domain)
	createDomain(spath, cpath, domain, pass)

}
