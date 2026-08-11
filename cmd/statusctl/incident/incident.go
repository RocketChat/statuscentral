package incident

import (
	"fmt"

	"github.com/spf13/cobra"
)

var SubCommands []*cobra.Command

var outputFormat = "list"

var IncidentCmd = &cobra.Command{
	Use: "incidents",
	Aliases: []string{
		"incident",
		"i",
	},
	Short:   "StatusCentral incidents",
	Example: "statusctl incidents [command]",
	Args: func(c *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("%v requires arguments", c.UseLine())
		}

		return nil
	},
}

func init() {
	getCmd.Flags().StringVarP(&outputFormat, "output", "o", "list", "output format")
	listCmd.Flags().BoolVarP(&latestOnly, "latest", "l", false, "Show latest only")
	
	// Add flags for create command
	createCmd.Flags().StringVar(&createTitle, "title", "", "Incident title")
	createCmd.Flags().StringVar(&createDescription, "description", "", "Incident description")
	createCmd.Flags().StringVar(&createEnvironment, "environment", "", "Environment label (e.g., dedicated-workspace-1)")
	createCmd.Flags().StringVar(&createDeployment, "deployment", "", "Deployment label (e.g., open)")

	// Add flags for patch command
	patchCmd.Flags().StringVar(&patchLabels, "labels", "", "Labels to patch in format: key1=value1,key2=value2")

	SubCommands = append(SubCommands, listCmd, describeCmd, getCmd, createCmd, patchCmd, updateCmd)
	IncidentCmd.AddCommand(SubCommands...)
}
