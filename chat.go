package main

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/code-to-go/safepool/apps/chat"
	registry "github.com/code-to-go/safepool/apps/registry"
	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/pool"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func printChatHelp() {
	color.White("commands: ")
	color.White("  '' refresh chat content")
	color.White("  \\x exit chat")
	color.White("  \\c create a sub pool")
}

var isValidName = regexp.MustCompile(`^[a-zA-Z0-9#]+$`).MatchString

const tokenContentType = "safepool/token"

func createChat(c chat.Chat) {

	var name, subject string
	for {
		prompt := promptui.Prompt{
			Label:       "Pool name (only alphanumeric and #): ",
			HideEntered: true,
		}

		name, _ = prompt.Run()
		if name == "" {
			return
		}
		if isValidName(name) {
			break
		}
		color.Red("Invalid name '%s'. Name can contain only alphanumeric letters and #.", name)

		prompt = promptui.Prompt{
			Label:       "Subject: ",
			HideEntered: true,
		}
		subject, _ = prompt.Run()
	}

	selfId := c.Pool.Self.Id()
	selected := map[string]bool{}
	for {
		items := []string{"Complete"}
		identities, _ := c.Pool.Identities()
		for idx, i := range identities {
			id := i.Id()
			if id == selfId {
				if idx < len(identities)-1 {
					identities[idx] = identities[len(identities)-1]
					identities[len(identities)-1] = i
				} else {
					continue
				}
			}
			if selected[id] {
				items = append(items, fmt.Sprintf("✓ %s [%s]", i.Nick, id))
			} else {
				items = append(items, fmt.Sprintf("  %s [%s]", i.Nick, id))
			}
		}

		sel := promptui.Select{
			Label: "Select users for the new pool",
			Items: items,
		}
		idx, _, err := sel.Run()
		if err != nil {
			return
		}

		if idx == 0 {
			break
		}
		id := identities[idx-1].Id()
		selected[id] = !selected[id]
	}

	var ids []string
	for id, ok := range selected {
		if ok {
			ids = append(ids, id)
		}
	}

	co, err := c.Pool.Sub(name, ids, c.Pool.Apps)
	if core.IsErr(err, "cannot create branch in pool %v: %v", c.Pool) {
		color.Red("😱 something went wrong!")
	}

	i := registry.Invite{
		Subject:      subject,
		Sender:       c.Pool.Self,
		Config:       &co,
		RecipientIds: ids,
	}
	tk, err := registry.Encode(i)
	if err == nil {
		c.SendMessage(tokenContentType, tk, nil)
	}
}

func processInvite(c chat.Chat, m chat.Message, invites []registry.Invite) []registry.Invite {
	invite, err := registry.Decode(c.Pool.Self, m.Text)
	if err == nil && invite.Config != nil {
		invites = append(invites, invite)
		color.Cyan("\t🔥 %s is inviting you to join %s; enter \\a to accept", invite.Sender.Nick, invite.Config.Name)
	}
	return invites
}

func acceptInvites(c chat.Chat, invites []registry.Invite) {
	items := []string{"cancel"}
	for _, i := range invites {
		items = append(items, fmt.Sprintf("%s by %s", i.Sender.Nick, i.Config.Name))
	}

	sel := promptui.Select{
		Label: "Select the pool you want to join",
		Items: items,
	}
	choice, _, err := sel.Run()
	if err != nil || choice == 0 {
		return
	}

	err = invites[choice-1].Join()
	if err == nil {
		color.Green("Congrats. You can now access to '%s'", invites[choice-1].Config.Name)
	} else {
		color.Red("Something went wrong: %v", err)
	}
}

func Chat(p *pool.Pool) {
	var lastId uint64
	var invites []registry.Invite
	c := chat.Get(p, "chat")

	identities, err := p.Identities()
	if err != nil {
		color.Red("cannot retrieve identities for pool '%s': %v", p.Name)
		return
	}

	id2nick := map[string]string{}
	for _, i := range identities {
		id2nick[i.Id()] = i.Nick
	}

	selfId := p.Self.Id()
	color.Green("Enter \\? for list of commands")
	for {
		messages, err := c.GetMessages(lastId, math.MaxInt64, 32)
		if err != nil {
			color.Red("cannot retrieve chat messages from pool '%s': %v", p.Name)
			return
		}
		for _, m := range messages {
			if m.ContentType == tokenContentType {
				invites = processInvite(c, m, invites)
				continue
			}

			if m.Author == selfId {
				color.Blue("%s: %s", id2nick[m.Author], m.Text)
			} else {
				color.Green("\t%s: %s", id2nick[m.Author], m.Text)
			}
			if m.Id > lastId {
				lastId = m.Id
			}
		}
		prompt := promptui.Prompt{
			Label:       "> ",
			HideEntered: true,
		}

		t, _ := prompt.Run()
		t = strings.Trim(t, " ")

		switch {
		case len(t) == 0:
		case strings.HasPrefix(t, "\\x"):
			return
		case strings.HasPrefix(t, "\\c"):
			createChat(c)
		case strings.HasPrefix(t, "\\a"):
			acceptInvites(c, invites)
		case strings.HasPrefix(t, "\\?"):
			printChatHelp()
		case strings.HasPrefix(t, "\\"):
			printChatHelp()
		default:
			_, err := c.SendMessage("text/html", t, nil)
			if err != nil {
				color.Red("cannot send message: %s")
			}
		}
	}
}
