package incident

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/RocketChat/statuscentral/cmd/statusctl/common"
	"github.com/spf13/cobra"
)

var (
	patchLabels string
)

var patchCmd = &cobra.Command{
	Use:   "patch <incident_id>",
	Short: "patch incident with labels",
	Example: `statusctl incident patch 123 --labels "environment=dedicated-workspace-1,deployment=open"`,
	Args:  cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		client := common.GetStatusCentralClient()

		incidentID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalln("Invalid incident ID:", args[0])
		}

		// Parse labels from string format "key1=value1,key2=value2"
		labels := make(map[string]string)
		if patchLabels != "" {
			pairs := strings.Split(patchLabels, ",")
			for _, pair := range pairs {
				kv := strings.Split(strings.TrimSpace(pair), "=")
				if len(kv) != 2 {
					log.Fatalln("Invalid labels format. Use: key1=value1,key2=value2")
				}
				labels[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}

		// Create a patch object with only labels
		patchData := struct {
			Labels map[string]string `json:"labels,omitempty"`
		}{
			Labels: labels,
		}

		// Use the client to patch the incident
		updatedIncident, err := client.Incidents().Patch(incidentID, patchData)
		if err != nil {
			panic(err)
		}

		log.Println(fmt.Sprintf("Incident %d patched successfully!", updatedIncident.ID))

		rendered, err := renderIncident(updatedIncident)
		if err != nil {
			panic(err)
		}

		log.Println(rendered)
	},
}