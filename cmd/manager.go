package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/surajsharma/kanastar/manager"
	"github.com/surajsharma/kanastar/utils"
)

// managerCmd represents the manager command
var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Manager command to operate a Kanastar manager node.",
	Long: `Kanastar manager command.

	The manager controls the orchestration system and is responsible for:
	◻ Accepting tasks from users
	◻ Scheduling tasks onto worker nodes
	◻ Rescheduling tasks in the event of a node failure
	◻ Periodically polling workers to get task updates`,

	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsDockerDaemonUp() {
			return
		}

		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		workers, _ := cmd.Flags().GetStringSlice("workers")
		scheduler, _ := cmd.Flags().GetString("scheduler")
		dbType, _ := cmd.Flags().GetString("dbType")

		log.Println("[cmd] starting manager")
		m := manager.New(workers, scheduler, dbType)
		api := manager.Api{Address: host, Port: port, Manager: m}
		go m.ProcessTasks()
		go m.UpdateTasks()
		go m.DoHealthChecks()
		// go m.UpdateNodeStats()
		log.Printf("[cmd] starting manager API on http://%s:%d", host, port)
		api.Start()

	},
}

func init() {
	rootCmd.AddCommand(managerCmd)
	managerCmd.Flags().StringP("host", "H", "0.0.0.0", "Hostname or IP address")
	managerCmd.Flags().IntP("port", "p", 5555, "Port on which to listen")
	managerCmd.Flags().StringSliceP("workers", "w", []string{"localhost:5556"}, "List (csv) of workers on which the manager will schedule tasks.")
	managerCmd.Flags().StringP("scheduler", "s", "epvm", "Name of scheduler to use (\"epvm\",\"roundrobin\", or \"greedy\")")
	managerCmd.Flags().StringP("dbType", "d", "memory", "Type of datastore to use for events and tasks (\"memory\" or \"persistent\")")

}
