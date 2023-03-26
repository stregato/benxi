package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/code-to-go/safepool/apps/chat"
	"github.com/code-to-go/safepool/apps/invite"

	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/pool"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

func printChatHelp() {
	color.White("commands: ")
	color.White("  CR refresh chat content")
	color.White("  double CR exit chat")
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
		identities, _ := c.Pool.Users()
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

	i := invite.Invite{
		Subject:      subject,
		Sender:       c.Pool.Self,
		Name:         co.Name,
		Exchanges:    co.Public,
		RecipientIds: ids,
	}
	tk, err := invite.Encode(i)
	if err == nil {
		c.SendMessage(tokenContentType, tk, nil, nil)
	}
}

func processInvite(c chat.Chat, m chat.Message, invites []invite.Invite) []invite.Invite {
	i, err := invite.Decode(c.Pool.Self, m.Text)
	if err == nil && i.Exchanges != nil {
		invites = append(invites, i)
		color.Cyan("\t🔥 %s is inviting you to join %s; enter \\a to accept", i.Sender.Nick, i.Name)
	}
	return invites
}

func acceptInvites(c chat.Chat, invites []invite.Invite) {
	items := []string{"cancel"}
	for _, i := range invites {
		items = append(items, fmt.Sprintf("%s by %s", i.Sender.Nick, i.Name))
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
		color.Green("Congrats. You can now access to '%s'", invites[choice-1].Name)
	} else {
		color.Red("Something went wrong: %v", err)
	}
}

const kitchen2 = "Mon 3:04PM"

func Chat(p *pool.Pool) {
	var after, before time.Time
	var invites []invite.Invite
	var private chat.Private
	recents := map[uint64]bool{}
	c := chat.Get(p, "chat")

	identities, err := p.Users()
	if err != nil {
		color.Red("cannot retrieve identities for pool '%s': %v", p.Name)
		return
	}

	id2nick := map[string]string{}
	nick2id := map[string]string{}
	for _, i := range identities {
		id2nick[i.Id()] = i.Nick
		nick2id[i.Nick] = i.Id()
	}

	lastCR := core.Now()
	before = core.Now()
	selfId := p.Self.Id()
	color.Green("Enter \\? for list of commands")
	for {
		messages, err := c.Receive(after, before, 32, private)
		if err != nil {
			color.Red("cannot retrieve chat messages from pool '%s': %v", p.Name)
			return
		}
		private = nil
		for _, m := range messages {
			if m.ContentType == tokenContentType {
				invites = processInvite(c, m, invites)
				continue
			}

			if m.Author == selfId {
				if !recents[m.Id] {
					color.Blue("\t%s 🕑%s: %s", id2nick[m.Author], m.Time.Format(kitchen2), m.Text)
				}
			} else {
				color.Green("%s 🕑%s: %s", id2nick[m.Author], m.Time.Format(kitchen2), m.Text)
			}
			after = m.Time
		}
		prompt := promptui.Prompt{
			Label:       "> ",
			HideEntered: true,
		}

		t, _ := prompt.Run()
		t = strings.Trim(t, " ")

		switch {
		case len(t) == 0:
			if core.Since(lastCR) < time.Second {
				return
			} else {
				lastCR = core.Now()
				before = core.Now()
			}
		case strings.HasPrefix(t, "\\c"):
			createChat(c)
		case strings.HasPrefix(t, "\\a"):
			acceptInvites(c, invites)
		case strings.HasPrefix(t, "\\?"):
			printChatHelp()
		case strings.HasPrefix(t, "\\"):
			printChatHelp()
		default:
			if t[0] == '|' {
				var nick string
				si := strings.Index(t, " ")
				if si != -1 {
					nick = t[0:si]
					t = strings.Trim(t[si:], " ")
				} else {
					nick = t[1:]
					t = ""
				}
				id := nick2id[nick]
				if id == "" {
					color.Red("unknown nick %s", nick)
					continue
				} else {
					private = []string{selfId, id}
				}
				if t == "" {
					continue
				}
			}

			id, err := c.SendMessage("text/html", t, nil, private)
			if err != nil {
				color.Red("cannot send message: %s")
			} else {
				color.Blue("%s 🕑%s: %s ", p.Self.Nick, core.Now().Format(kitchen2), t)
				recents[id] = true
			}
		}
	}
}
