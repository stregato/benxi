package main

import (
	"github.com/code-to-go/safepool/api"
	"github.com/code-to-go/safepool/apps/invite"
	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/pool"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func CreateInvite(p *pool.Pool) {
	prompt := promptui.Prompt{
		Label: "Enter id (or empty for global token)",
	}
	id, _ := prompt.Run()

	token, err := api.PoolInvite(p.Name, []string{id}, "")
	if core.IsErr(err, "cannot generate invite: %v") {
		color.Red("cannot generate invite: %v", err)
		return
	}
	if id == "" {
		color.Green("Universal token:\n%s", token)
	} else {
		color.Green("Token for id '%s'\n%s", id, token)
	}
}

func Invites(p *pool.Pool) {

	for {
		identities, err := p.Users()
		if core.IsErr(err, "cannot read identities from db: %v") {
			color.Red("cannot list users")
			return
		}

		color.Green("Users")
		for _, i := range identities {
			color.Green("\t%s %s - %s", i.Nick, i.Email, i.Id())
		}

		items := []string{"Action: Invite", "Action: Sub Pool", "Action: Back"}
		prompt := promptui.Select{
			Label: "Choose",
			Items: items,
		}

		color.Green("Invites")
		invites, _ := invite.Receive(p, 0, false)
		for _, i := range invites {
			if i.Exchanges == nil {
				color.Green("%s has sent an invite", i.Sender.Nick)
			} else {
				items = append(items, "Accept invite to '%s' by '%s' ")
			}
		}

		idx, _, _ := prompt.Run()
		switch idx {
		case 0:
			CreateInvite(p)
		case 2:
			return
		default:
		}

	}
}
