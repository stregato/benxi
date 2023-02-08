package main

import (
	"fmt"

	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/security"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func Trust() {
	for {
		items := []string{"Action: Back"}

		identities, err := security.Identities()
		if core.IsErr(err, "cannot read identities: %v") {
			color.Red("cannot read identities")
			return
		}

		trustedSet := map[string]bool{}
		trusted, _ := security.Trusted()
		for _, t := range trusted {
			trustedSet[t.Id()] = true
		}

		for _, i := range identities {
			if trustedSet[i.Id()] {
				items = append(items, fmt.Sprintf("%s (T) %s - %s", i.Nick, i.Email, i.Id()))
			} else {
				items = append(items, fmt.Sprintf("%s ( ) %s - %s", i.Nick, i.Email, i.Id()))
			}
		}

		prompt := promptui.Select{
			Label: "Select to trust/untrust a bather",
			Items: items,
		}

		idx, _, _ := prompt.Run()
		if idx == 0 {
			return
		}

		identity := identities[idx-1]
		security.Trust(identity, !trustedSet[identity.Id()])
		trustedSet[identity.Id()] = !trustedSet[identity.Id()]
	}

}
