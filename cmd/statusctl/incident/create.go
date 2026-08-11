package incident

import (
	"errors"
	"fmt"
	"log"

	"github.com/RocketChat/statuscentral/cmd/statusctl/common"
	"github.com/RocketChat/statuscentral/models"
	"github.com/spf13/cobra"
)

var (
	createTitle       string
	createDescription string
	createEnvironment string
	createDeployment  string
)

var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "create incident",
	Example: "statusctl incident create --title \"Database Issue\" --description \"Connectivity issue\" --environment \"dedicated-workspace-1\" --deployment \"open\"",
	Run: func(c *cobra.Command, args []string) {
		client := common.GetStatusCentralClient()

		services, err := client.Services().GetMultiple()
		if err != nil {
			panic(err)
		}

		var title, environment, deployment string

		// Get title either from flag or prompt
		if createTitle != "" {
			title = createTitle
		} else {
			title = common.StringPrompt("Incident Short Description / Title:")
		}

		// Get environment either from flag or prompt
		if createEnvironment != "" {
			environment = createEnvironment
		} else {
			environment = common.StringPromptWithDefault("Environment (e.g., dedicated-workspace-1):", "")
		}

		// Get deployment either from flag or prompt
		if createDeployment != "" {
			deployment = createDeployment
		} else {
			deployment = common.StringPromptWithDefault("Deployment (e.g., open):", "")
		}

		for i, statusOption := range models.IncidentStatusArray {
			log.Printf("%d) %s\n", i, statusOption)
		}

		status, err := common.IntPrompt("Current Incident Status [1]:", 1)
		if err != nil {
			log.Fatalln("Invalid selection")
		}

		servicesImpacted, err := getImpactedServices(services)
		if err != nil {
			panic(err)
		}

		incident := &models.Incident{
			Title:    title,
			Status:   models.IncidentStatusArray[status],
			Services: servicesImpacted,
		}

		// Add labels if provided
		if environment != "" || deployment != "" {
			incident.Labels = make(map[string]string)
			if environment != "" {
				incident.Labels["environment"] = environment
			}
			if deployment != "" {
				incident.Labels["deployment"] = deployment
			}
		}

		returnedIncident, err := client.Incidents().Create(incident)
		if err != nil {
			panic(err)
		}

		log.Println(fmt.Sprintf("Incident %d created!", returnedIncident.ID))

		rendered, err := renderIncident(returnedIncident)
		if err != nil {
			panic(err)
		}

		log.Println(rendered)
	},
}

func getImpactedServices(services []*models.Service) ([]models.ServiceUpdate, error) {
	gettingServices := true
	serviceUpdates := []models.ServiceUpdate{}

	for gettingServices {
		for i, service := range services {
			log.Printf("%d) %s\n", i, service.Name)
		}

		service, err := common.IntPrompt("Select a service impacted [1]:", 1)
		if err != nil {
			return serviceUpdates, errors.New("invalid selection")
		}

		for i, serviceStatus := range models.ServiceStatusArray {
			log.Printf("%d) %s\n", i, serviceStatus)
		}

		serviceStatus, err := common.IntPrompt("Select service status [1]:", 1)
		if err != nil {
			return serviceUpdates, errors.New("invalid selection")
		}

		serviceUpdates = append(serviceUpdates, models.ServiceUpdate{
			Name:    services[service].Name,
			Status:  models.ServiceStatusArray[serviceStatus],
			Regions: []string{},
		})

		more, err := common.GetYesNoPrompt("Add another service?", true)
		if err != nil {
			return serviceUpdates, err
		}

		if !more {
			gettingServices = false
		}
	}

	return serviceUpdates, nil
}
