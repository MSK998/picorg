/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/MSK998/picorg/app"
	"github.com/spf13/cobra"
)

var App *app.App

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "picorg [path]",
	Short: "Will sort a folder based on dates noted in files",
	Long: `
picorg will organise based on dates of file names.
It can be expanded to look at other ways to sort from metadata.
	
It will walk the directory noted in the first arg, 
and based on regex in golang syntax will sort it based on the 
groups that have been set within the regex defined.
	`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		drFlag, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		dir, err := filepath.Abs(args[0])
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(2)
		}

		err = App.Picorg(&app.Picorg{
			IsDryRun: drFlag,
			Dir: dir,
			Regex: regexp.MustCompile(`^(\d{2})-(\d{2})`),
		})
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(3)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	var err error
	App, err = app.New()
	if err != nil {
		panic(err.Error())
	}


	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.picorg.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("dry-run", "d", false, "Dry-run to show how the folder structure will look after")
}


