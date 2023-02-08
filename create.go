package main

import (
	"os"
	"path/filepath"

	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/pool"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v3"
)

func Create() {

	for {
		files, err := os.ReadDir(".")
		if core.IsErr(err, "cannot list current folder: %v") {
			color.Red("internal error")
			continue
		}

		items := []string{"Back"}
		for _, f := range files {
			if filepath.Ext(f.Name()) == ".yaml" {
				items = append(items, f.Name())
			}
		}

		prompt := promptui.Select{
			Label: "Select the configuration file (yaml)",
			Items: items,
		}

		idx, _, _ := prompt.Run()
		if idx == 0 {
			return
		}

		data, err := os.ReadFile(items[idx])
		if core.IsErr(err, "cannot read file '%s': %v", items[idx]) {
			color.Red("cannot read file '%s': %v", items[idx], err)
			continue
		}

		color.Green("Config file\n--\n%s\n--", string(data))
		var c pool.Config
		err = yaml.Unmarshal(data, &c)
		if core.IsErr(err, "invalid config: %v") {
			color.Red("invalid config: %v", err)
			continue
		}

		err = pool.Define(c)
		if core.IsErr(err, "cannot define config: %v") {
			color.Red("internal error in setting the config pool")
			continue
		}

		p, err := pool.Create(safepool.Self, c.Name, safepool.Apps)
		if core.IsErr(err, "cannot create pool: %v") {
			color.Red("Cannot create pool: %v", err)
			continue
		}
		p.Close()
		color.Green("Pool %s created", c.Name)
	}

}
