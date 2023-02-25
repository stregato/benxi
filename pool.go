package main

import (
	"github.com/code-to-go/safepool/api"
	"github.com/code-to-go/safepool/pool"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func ChooseFunction(poolName string) {
	p, err := pool.Open(api.Self, poolName)
	if err != nil {
		color.Red("cannot open pool '%s': %v", poolName, err)
		return
	}
	defer p.Close()

	for {
		items := []string{"🗨 Chat", "📚 Library", "📧 Invites", "🗑 Leave", "🔙 Back"}
		prompt := promptui.Select{
			Label: "Choose App",
			Items: items,
		}

		idx, _, _ := prompt.Run()
		switch idx {
		case 0:
			Chat(p)
		case 1:
			Library(p)
		case 2:
			Invites(p)
		case 3:
			if api.PoolLeave(poolName) == nil {
				return
			}
		default:
			return
		}
	}
}

func SelectMain() {
	for {
		var items []string
		pools := pool.List()
		for _, p := range pools {
			items = append(items, "🌊 "+p)
		}
		items = append(items, []string{"＋ Add Pool", "🆕 Create Pool", "👤 Trust User", "⚙ Settings", "✖ Exit"}...)
		prompt := promptui.Select{
			Label: "Choose",
			Items: items,
			Size:  20,
		}

		idx, _, _ := prompt.Run()
		switch idx {
		case len(pools):
			AddExisting()
		case len(pools) + 1:
			Create()
		case len(pools) + 2:
			Trust()
		case len(pools) + 3:
			Settings()
		case len(pools) + 4:
			return
		default:
			ChooseFunction(pools[idx])
		}
	}
}

func Settings() {
	color.Green("My Nick: %s", api.Self.Nick)
	color.Green("My Public id: %s", api.Self.Id())
}
