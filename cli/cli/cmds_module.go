package cli

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (c *CLI) initModuleCmdsAndFlags() {
	c.initModuleCmds()
	c.initModuleFlags()
}

func (c *CLI) initModuleCmds() {
	c.moduleCmd = &cobra.Command{
		Use:   "module",
		Short: "The module manager",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
	c.serviceCmd.AddCommand(c.moduleCmd)

	c.moduleTypesCmd = &cobra.Command{
		Use:   "types",
		Short: "List the available module types and their IDs",
		Run: func(cmd *cobra.Command, args []string) {

			client := newHTTPClient()
			const u = "http://s/r/module/types"

			resp, respErr := client.Get(u)
			if respErr != nil {
				panic(respErr)
			}

			defer resp.Body.Close()
			body, bodyErr := ioutil.ReadAll(resp.Body)
			if bodyErr != nil {
				panic(bodyErr)
			}

			fmt.Println(string(body))
		},
	}
	c.moduleCmd.AddCommand(c.moduleTypesCmd)

	c.moduleInstancesCmd = &cobra.Command{
		Use:   "instance",
		Short: "The module instance manager",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
	c.moduleCmd.AddCommand(c.moduleInstancesCmd)

	c.moduleInstancesListCmd = &cobra.Command{
		Use:     "get",
		Aliases: []string{"ls", "list"},
		Short:   "List the running module instances",
		Run: func(cmd *cobra.Command, args []string) {

			client := newHTTPClient()
			const u = "http://s/r/module/instances"

			resp, respErr := client.Get(u)
			if respErr != nil {
				panic(respErr)
			}

			defer resp.Body.Close()
			body, bodyErr := ioutil.ReadAll(resp.Body)
			if bodyErr != nil {
				panic(bodyErr)
			}

			fmt.Println(string(body))
		},
	}
	c.moduleInstancesCmd.AddCommand(c.moduleInstancesListCmd)

	c.moduleInstancesCreateCmd = &cobra.Command{
		Use:     "create",
		Aliases: []string{"new"},
		Short:   "Create a new module instance",
		Run: func(cmd *cobra.Command, args []string) {

			client := newHTTPClient()
			const u = "http://s/r/module/instances"

			if c.moduleTypeName == "" || c.moduleInstanceAddress == "" {
				cmd.Usage()
				return
			}

			modInstStartStr := fmt.Sprintf("%v", c.moduleInstanceStart)

			cfgJSON, cfgJSONErr := c.config.ToJSON()

			if cfgJSONErr != nil {
				panic(cfgJSONErr)
			}

			log.WithFields(log.Fields{
				"url":      u,
				"name":     c.moduleInstanceName,
				"typeName": c.moduleTypeName,
				"address":  c.moduleInstanceAddress,
				"start":    modInstStartStr,
				"config":   cfgJSON}).Debug("post create module instance")

			resp, respErr := client.PostForm(u,
				url.Values{
					"name":     {c.moduleInstanceName},
					"typeName": {c.moduleTypeName},
					"address":  {c.moduleInstanceAddress},
					"start":    {modInstStartStr},
					"config":   {cfgJSON},
				})
			if respErr != nil {
				panic(respErr)
			}

			defer resp.Body.Close()
			body, bodyErr := ioutil.ReadAll(resp.Body)
			if bodyErr != nil {
				panic(bodyErr)
			}

			fmt.Println(string(body))
		},
	}
	c.moduleInstancesCmd.AddCommand(c.moduleInstancesCreateCmd)

	c.moduleInstancesStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Starts a module instance",
		Run: func(cmd *cobra.Command, args []string) {

			if c.moduleInstanceName == "" {
				cmd.Usage()
				return
			}

			client := newHTTPClient()
			u := fmt.Sprintf(
				"http://s/r/module/instances/%s/start", c.moduleInstanceName)

			resp, respErr := client.Get(u)
			if respErr != nil {
				panic(respErr)
			}

			defer resp.Body.Close()
			body, bodyErr := ioutil.ReadAll(resp.Body)
			if bodyErr != nil {
				panic(bodyErr)
			}

			fmt.Println(string(body))
		},
	}
	c.moduleInstancesCmd.AddCommand(c.moduleInstancesStartCmd)
}

func (c *CLI) initModuleFlags() {
	c.moduleInstancesCreateCmd.Flags().StringVarP(&c.moduleTypeName, "typeName",
		"t", "", "The name of the module type to instance")

	c.moduleInstancesCreateCmd.Flags().StringVarP(&c.moduleInstanceName, "name",
		"n", "", "The name of the new module instance")

	c.moduleInstancesCreateCmd.Flags().StringVarP(&c.moduleInstanceAddress,
		"address", "a", "",
		"The network address at which the module will be exposed")

	c.moduleInstancesCreateCmd.Flags().BoolVarP(&c.moduleInstanceStart,
		"start", "s", false,
		"A flag indicating whether or not to start the module upon creation")

	c.moduleInstancesCreateCmd.Flags().StringSliceVarP(&c.moduleConfig,
		"options", "o", nil,
		"A comma-seperated string of key=value pairs used by some module "+
			"types for custom configuraitons.")

	c.moduleInstancesStartCmd.Flags().StringVarP(&c.moduleInstanceName, "name",
		"n", "", "The name of the module instance to start")
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: func(string, string) (net.Conn, error) {
				return net.Dial("unix", serverSockFile)
			},
		},
	}
}
