package main

import (
	"fmt"
	"os"

	"github.com/708u/gwt"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gwt configuration",
	Long:  `Create a .gwt/settings.toml configuration file in the current directory.`,
	Args:  cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Override parent's PersistentPreRunE to skip config loading
		// since init creates the config file
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		cwd, err = resolveDirectory(dirFlag, cwd)
		if err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")

		initCommand := gwt.NewInitCommand()
		result, err := initCommand.Run(cwd, gwt.InitOptions{Force: force})
		if err != nil {
			return err
		}

		formatted := result.Format(gwt.InitFormatOptions{})
		fmt.Fprint(os.Stdout, formatted.Stdout)
		return nil
	},
}

func init() {
	initCmd.Flags().BoolP("force", "f", false, "Overwrite existing configuration file")
	rootCmd.AddCommand(initCmd)
}
