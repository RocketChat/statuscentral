package incident

import (
	"bytes"
	"html/template"
	"log"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/RocketChat/statuscentral/cmd/statusctl/common"
	"github.com/RocketChat/statuscentral/models"
)

var incidentDetailTemplate = `
ID: {{.ID}}
Title: {{.Title}}
Created: {{ .Time.Format "Jan 02 2006 15:04" }}
Status: {{.Status}}
{{ if .Labels }}Labels:
{{ range $key, $value := .Labels }}- {{$key}}: {{$value}}
{{ end }}{{ end }}Services: 
{{ range $service := .Services }}
- Name: {{$service.Name}}
  Status: {{$service.Status}}
  Regions:
  {{ range $region := $service.Regions }}
  - 
  {{ end }}
{{ end }}

Updates:
{{ range $update := .Updates }}
- ID: {{$update.ID}}
  Time: {{ $update.Time.Format "Jan 02 2006 15:04" }}
  Status: {{ $update.Status }}
  Message: {{ $update.Message }}
{{ end }}
`

var describeCmd = &cobra.Command{
	Use:     "describe",
	Short:   "describe incident",
	Example: "statusctl incident describe",
	Run: func(c *cobra.Command, args []string) {
		client := common.GetStatusCentralClient()

		id, err := strconv.Atoi(args[0])
		if err != nil {
			panic("Unable to parse incident id")
		}

		incident, err := client.Incidents().Get(id)
		if err != nil {
			panic(err)
		}

		rendered, err := renderIncident(incident)
		if err != nil {
			panic(err)
		}

		log.Println(rendered)

	},
}

func renderIncident(incident *models.Incident) (string, error) {
	temp := template.New("incidentDetail")
	t, err := temp.Parse(incidentDetailTemplate)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	if err := t.Execute(buf, incident); err != nil {
		return "", err
	}

	return buf.String(), nil
}
