package main

import (
	"github.com/code-to-go/safepool/apps/invite"
	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/pool"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func CreateInvite(p *pool.Pool) {

	propmt := promptui.Prompt{
		Label: "Enter id (or empty for global token)",
	}
	id, _ := propmt.Run()

	c, err := pool.GetConfig(p.Name)
	if core.IsErr(err, "cannot read config for '%s': %v", p.Name) {
		color.Red("invalid config")
		return
	}
	i := invite.Invite{
		Config: &c,
		Sender: p.Self,
	}

	if id != "" {
		err = p.SetAccess(id, pool.Active)
		if core.IsErr(err, "cannot set access for id '%s' in pool '%s': %v", id, p.Name) {
			color.Red("id '%s' has some problems: %v", id, err)
			return
		}
		i.RecipientIds = append(i.RecipientIds, id)
	}

	token, err := invite.Encode(i)
	if core.IsErr(err, "cannot create token: %v") {
		color.Red("cannot create token: %v", err)
		return
	}

	invite.Add(p, i)
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
		invites, _ := invite.GetInvites(p, 0, false)
		for _, i := range invites {
			if i.Config == nil {
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
