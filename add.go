package main

import (
	"github.com/code-to-go/safepool/api"
	"github.com/code-to-go/safepool/apps/invite"
	"github.com/code-to-go/safepool/core"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func AddExisting() {
	color.Green("My Public id: %s", api.Self.Id())

	for {
		prompt := promptui.Prompt{
			Label:       "Invite",
			HideEntered: true,
		}

		t, _ := prompt.Run()
		if len(t) == 0 {
			return
		}

		i, err := invite.Decode(api.Self, t)
		if core.IsErr(err, "invalid token: %v") {
			continue
		}

		if i.Exchanges == nil {
			color.Red("the invite is not for you")
			continue
		}

		if core.IsErr(i.Join(), "cannot join the pool: %s") {
			continue
		}

		color.Green("Pool %s added. Host '%s' is trusted", i.Name, i.Sender.Nick)
		return
	}

}

func AddPool() {

	items := []string{"Add Existing", "Create New", "Cancel"}
	prompt := promptui.Select{
		Label: "Choose",
		Items: items,
	}

	idx, _, _ := prompt.Run()
	switch idx {
	case 0:
	}

}
