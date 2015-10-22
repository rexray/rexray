package cli

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/emccode/rexray/util"
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

			_, addr, addrErr := util.ParseAddress(c.r.Config.Host)
			if addrErr != nil {
				panic(addrErr)
			}

			u := fmt.Sprintf("http://%s/r/module/types", addr)

			client := &http.Client{}
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

			_, addr, addrErr := util.ParseAddress(c.r.Config.Host)
			if addrErr != nil {
				panic(addrErr)
			}

			u := fmt.Sprintf("http://%s/r/module/instances", addr)

			client := &http.Client{}
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

			_, addr, addrErr := util.ParseAddress(c.r.Config.Host)
			if addrErr != nil {
				panic(addrErr)
			}

			if c.moduleTypeID == -1 || c.moduleInstanceAddress == "" {
				cmd.Usage()
				return
			}

			modTypeIDStr := fmt.Sprintf("%d", c.moduleTypeID)
			modInstStartStr := fmt.Sprintf("%v", c.moduleInstanceStart)

			u := fmt.Sprintf("http://%s/r/module/instances", addr)
			cfgJSON, cfgJSONErr := c.r.Config.ToJSON()

			if cfgJSONErr != nil {
				panic(cfgJSONErr)
			}

			log.WithFields(log.Fields{
				"url":     u,
				"typeId":  modTypeIDStr,
				"address": c.moduleInstanceAddress,
				"start":   modInstStartStr,
				"config":  cfgJSON}).Debug("post create module instance")

			client := &http.Client{}
			resp, respErr := client.PostForm(u,
				url.Values{
					"typeId":  {modTypeIDStr},
					"address": {c.moduleInstanceAddress},
					"start":   {modInstStartStr},
					"config":  {cfgJSON},
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

			_, addr, addrErr := util.ParseAddress(c.r.Config.Host)
			if addrErr != nil {
				panic(addrErr)
			}

			if c.moduleInstanceID == -1 {
				cmd.Usage()
				return
			}

			u := fmt.Sprintf(
				"http://%s/r/module/instances/%d/start", addr, c.moduleInstanceID)

			client := &http.Client{}
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
	c.moduleInstancesCreateCmd.Flags().Int32VarP(&c.moduleTypeID, "id",
		"i", -1, "The ID of the module type to instance")

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

	c.moduleInstancesStartCmd.Flags().Int32VarP(&c.moduleInstanceID, "id",
		"i", -1, "The ID of the module instance to start")
}
