package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ehrlich-b/cube/internal/web"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web server",
	Long: `Start the web server to provide a browser-based interface
for the cube solver.`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")
		host, _ := cmd.Flags().GetString("host")
		
		fmt.Printf("Starting web server at http://%s:%s\n", host, port)
		
		server := web.NewServer()
		if err := server.Start(host + ":" + port); err != nil {
			fmt.Printf("Error starting server: %v\n", err)
		}
	},
}

func init() {
	serveCmd.Flags().StringP("port", "p", "8080", "Port to run the server on")
	serveCmd.Flags().StringP("host", "H", "localhost", "Host to bind the server to")
}