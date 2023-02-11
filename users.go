package main

import (
	"fmt"

	"github.com/code-to-go/safepool/apps/registry"
	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/pool"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func AddUser(p *pool.Pool) {

	propmt := promptui.Prompt{
		Label: "Enter id (or empty for global token)",
	}
	id, _ := propmt.Run()

	c, err := pool.GetConfig(p.Name)
	if core.IsErr(err, "cannot read config for '%s': %v", p.Name) {
		color.Red("invalid config")
		return
	}
	t := registry.Invite{
		Config: &c,
		Sender: p.Self,
	}

	if id != "" {
		err = p.SetAccess(id, pool.Active)
		if core.IsErr(err, "cannot set access for id '%s' in pool '%s': %v", id, p.Name) {
			color.Red("id '%s' has some problems: %v", id, err)
			return
		}
		t.RecipientIds = append(t.RecipientIds, id)
	}

	token, err := registry.Encode(t)
	if core.IsErr(err, "cannot create token: %v") {
		color.Red("cannot create token: %v", err)
		return
	}

	if id == "" {
		color.Green("Universal token:\n%s", token)
	} else {
		color.Green("Token for id '%s'\n%s", id, token)
	}
}

func Users(p *pool.Pool) {

	for {

		identities, err := p.Identities()
		if core.IsErr(err, "cannot read identities from db: %v") {
			color.Red("cannot list users")
			return
		}

		items := []string{"Action: Add", "Action: Back"}
		for _, i := range identities {
			items = append(items, fmt.Sprintf("%s %s - %s", i.Nick, i.Email, i.Id()))
		}

		prompt := promptui.Select{
			Label: "Choose",
			Items: items,
		}

		idx, _, _ := prompt.Run()
		switch idx {
		case 0:
			AddUser(p)
		case 1:
			return
		default:
		}

	}

}
