package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("manager", "m", "localhost:5555", "Manager to talk to")
	runCmd.Flags().StringP("filename", "f", "task.json", "Task specification file")
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)

	return !errors.Is(err, fs.ErrNotExist)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a new task.",
	Long: `Kanastar run command.

The run command starts a new task.`,
	Run: func(cmd *cobra.Command, args []string) {

		manager, _ := cmd.Flags().GetString("manager")
		filename, _ := cmd.Flags().GetString("filename")

		fullFilePath, err := filepath.Abs(filename)
		if err != nil {
			log.Fatal(err)
		}

		if !fileExists(fullFilePath) {
			log.Fatalf("[cmd] file %s does not exist.", filename)
		}

		log.Printf("[cmd] using manager: %v\n", manager)
		log.Printf("[cmd] using file: %v\n", fullFilePath)

		data, err := os.ReadFile(filename)
		if err != nil {
			log.Fatalf("[cmd] unable to read file: %v", filename)
		}

		log.Printf("\n[cmd] attempting to run task: %v\n", string(data))

		url := fmt.Sprintf("http://%s/tasks", manager)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Panic(err)
		}

		if resp.StatusCode != http.StatusCreated {
			log.Printf("[cmd] error sending request: %v", resp.StatusCode)
		}

		defer resp.Body.Close()
		log.Println("[cmd] successfully sent task request to manager")
	},
}
