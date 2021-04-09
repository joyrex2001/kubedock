package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/joyrex2001/donk/internal"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "donk",
	Short: "donk is a docker on kubernetes service.",
	Long:  ``,
	Run:   internal.Main,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().String("listen-addr", ":8080", "Webserver listen address")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose mode")
	viper.BindPFlag("generic.verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("server.listen-addr", rootCmd.PersistentFlags().Lookup("listen-addr"))
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		// fmt.Printf("not using config file: %s\n", err)
	} else {
		fmt.Printf("using config: %s\n", viper.ConfigFileUsed())
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
