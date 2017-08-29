package vault

import (
	"bufio"
	"os"
	"runtime"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/vault/api"
	cli "gopkg.in/urfave/cli.v1"
)

// FindToken ...
func FindToken(c *cli.Context) error {
	client, err := api.NewClient(nil)
	if err != nil {
		return err
	}

	response, err := client.Logical().List("/auth/token/accessors")
	if err != nil {
		return err
	}

	tokenReader := client.Auth().Token()

	accessors := getTokenAccessorsFromResponse(response.Data["keys"])
	log.Infof("Found %d possible tokens, beging scanning each ...", len(accessors))
	log.Info("")

	inCh := make(chan string, 9999)
	outCh := make(chan *api.Secret, 9999)

	wg := sync.WaitGroup{}
	wg.Add(len(accessors))

	for i := 0; i < runtime.NumCPU(); i++ {
		go func(i int) {
			log.Debugf("Starting indexer %d", i)

			for {
				select {
				case accessorString := <-inCh:
					token, err := tokenReader.LookupAccessor(accessorString)
					if err != nil {
						log.Errorf("Could not lookup accessor %s: %s", accessorString, err)
						wg.Done()
						continue
					}

					if filterToken(token.Data, c) {
						wg.Done()
						continue
					}

					outCh <- token
				}
			}
		}(i)
	}

	go func() {
		for _, accessorString := range accessors {
			inCh <- accessorString
		}
	}()

	go func() {
		for {
			select {
			case token := <-outCh:
				log.Infof("Found token: %s", token.Data["display_name"])
				log.Infof("  policies           : %s", getPolicies(token.Data))
				log.Infof("  orphan             : %t", token.Data["orphan"])
				log.Infof("  renewable          : %t", token.Data["renewable"])
				log.Infof("  path               : %s", token.Data["path"])
				log.Infof("  creation_time      : %s", token.Data["creation_time"])
				log.Infof("  ttl                : %s", token.Data["ttl"])
				log.Infof("  creation_ttl       : %s", token.Data["creation_ttl"])
				log.Infof("  explicit_max_ttl   : %s", token.Data["explicit_max_ttl"])
				log.Infof("  num_uses           : %s", token.Data["num_uses"])

				if meta, ok := token.Data["meta"]; ok {
					for k, v := range meta.(map[string]interface{}) {
						log.Infof("  meta[%s] %s: %v", k, strings.Repeat(" ", max(0, 12-len(k))), v)
					}
				}

				log.Info("")

				if c.Bool("delete-matches") {
					if confirm("Are you sure you want to delete this token?") {
						if err := tokenReader.RevokeAccessor(token.Data["accessor"].(string)); err != nil {
							log.Errorf("Could not delete token: %s", err)
						}
						log.Info("Successfully deleted token")
					} else {
						log.Info("Skipping delete")
					}
				}

				wg.Done()
			}
		}
	}()

	log.Info("Waiting ...")
	wg.Wait()

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func confirm(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		log.Warnf("%s: [y|n] ", message)
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Could not read stdin")
		}
		input = strings.Trim(input, "\n")
		input = strings.ToLower(input)

		switch input {
		case "y":
			return true
		case "n":
			return false
		default:
			log.Error("Expected 'y' or 'n', please try again")
		}
	}

}

func filterToken(data map[string]interface{}, c *cli.Context) bool {
	// filer on the name / display_name
	if name := c.String("filter-name"); name != "" && !strings.Contains(data["display_name"].(string), name) {
		return true
	}

	// filter on orphan token
	if oprhan := c.Bool("filter-orphan"); oprhan && !data["orphan"].(bool) {
		return true
	}

	// filter on policy for the token
	if policy := c.String("filter-policy"); policy != "" && !contains(getPolicies(data), policy) {
		return true
	}

	// filter on path for the token
	if path := c.String("filter-path"); path != "" && !strings.Contains(data["path"].(string), path) {
		return true
	}

	// filter on meta key "username"
	if username := c.String("filter-meta-username"); username != "" && username != getMetaKey(data, "username") {
		return true
	}

	return false
}

func getMetaKey(data map[string]interface{}, key string) string {
	metaRaw, ok := data["meta"]
	if !ok || metaRaw == nil {
		return ""
	}

	meta := metaRaw.(map[string]interface{})
	keyValueRaw, ok := meta[key]
	if !ok || keyValueRaw == nil {
		return ""
	}

	return keyValueRaw.(string)
}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func getPolicies(data map[string]interface{}) []string {
	out := make([]string, 0)
	old := data["policies"].([]interface{})
	for _, policy := range old {
		out = append(out, policy.(string))
	}

	return out
}

func getTokenAccessorsFromResponse(in interface{}) []string {
	old := in.([]interface{})

	out := make([]string, 0)
	for _, v := range old {
		out = append(out, v.(string))
	}

	return out
}
