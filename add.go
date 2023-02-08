package main

import (
	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/pool"
	"github.com/code-to-go/safepool/security"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func AddExisting() {

	for {
		prompt := promptui.Prompt{
			Label:       "Token from your host",
			HideEntered: true,
		}

		t, _ := prompt.Run()
		if len(t) == 0 {
			return
		}

		token, err := pool.DecodeToken(safepool.Self, t)
		if core.IsErr(err, "invalid token: %v") {
			continue
		}

		err = security.SetIdentity(token.Host)
		if core.IsErr(err, "cannot save identity '%s': %v", token.Host.Nick) {
			continue
		}

		err = security.Trust(token.Host, true)
		if core.IsErr(err, "cannot trust identity '%s': %v", token.Host.Nick) {
			continue
		}

		if core.IsErr(pool.Define(token.Config), "cannot save pool in db: %s") {
			continue
		}

		color.Green("Pool %s added. Host '%s' is trusted", token.Config.Name, token.Host.Nick)
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
