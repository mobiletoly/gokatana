package katapp

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v0.1")
	},
}

type CmdlineHandler struct {
	Run func(deployment string)
}

func CmdlineExecute(name, short, long string, hdl *CmdlineHandler) {
	var deployment string
	serverCmd := &cobra.Command{
		Use:   "run --deployment={local|dev|prod|...}",
		Short: "Run " + name,
		Long: "Run " + name + " and use configuration settings specified in 'deployment' flag. " +
			"Name of deployment corresponds to filename in config directory, e.g. 'run --deployment=local' " +
			"means that service will be started with config values loaded from configs/local.yaml file",
		Run: func(cmd *cobra.Command, args []string) {
			hdl.Run(deployment)
		},
	}
	serverCmd.Flags().StringVar(&deployment, "deployment", "",
		"deployment environment for API server, e.g. local, prod (it should match your config filename)")
	_ = serverCmd.MarkFlagRequired("deployment")

	var rootCmd = newCobraCmdlineCommand(name, short, long)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(serverCmd)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// newCobraCmdlineCommand is a helper function to add new command-line command and parameters
func newCobraCmdlineCommand(use string, short string, long string) cobra.Command {
	return cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			bindViperToCobraCommands([]*cobra.Command{cmd})
		},
	}
}

func bindViperToCobraCommands(commands []*cobra.Command) {
	for _, cmd := range commands {
		copyViperToCobraFlags(cmd)
		if cmd.HasSubCommands() {
			bindViperToCobraCommands(cmd.Commands())
		}
	}
}

func copyViperToCobraFlags(cmd *cobra.Command) {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		panic(err)
	}
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			if err := cmd.Flags().Set(f.Name, viper.GetString(f.Name)); err != nil {
				panic(err)
			}
		}
	})
}
