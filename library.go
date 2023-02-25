package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/code-to-go/safepool/apps/library"
	"github.com/code-to-go/safepool/core"
	"github.com/code-to-go/safepool/pool"
	"github.com/code-to-go/safepool/security"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/skratchdot/open-golang/open"
)

func addDocument(l library.Library) {
	for {
		prompt := promptui.Prompt{
			Label: "Local Path",
		}

		localPath, _ := prompt.Run()
		if localPath == "" {
			return
		}

		stat, err := os.Stat(localPath)
		if err != nil {
			color.Red("invalid path '%s'", localPath)
			continue
		}
		if stat.IsDir() {
			color.Red("folders are not supported at the moment")
			continue
		}

		var items []string
		var item string
		parts := strings.Split(localPath, string(os.PathSeparator))
		sort.Slice(parts, func(i, j int) bool { return i > j })
		for _, p := range parts {
			if p != "" {
				item = path.Join(p, item)
				items = append(items, item)
			}
		}

		sel := promptui.Select{
			Label: "Name in the pool",
			Items: items,
		}
		_, name, _ := sel.Run()

		prompt = promptui.Prompt{
			Label:   "Edit Name",
			Default: name,
		}
		name, _ = prompt.Run()
		h, err := l.Send(localPath, name, true)
		if core.IsErr(err, "cannot upload %s: %v", localPath) {
			color.Red("cannot upload %s", localPath)
		} else {
			color.Green("'%s' uploaded to '%s:%s' with id %d", localPath, l.Pool.Name, name, h.Id)
			return
		}
	}

}

type actionType int

const (
	uploadLocal actionType = iota
	openLocally
	openFolder
	deletelocal
	updateLocal
	downloadTemp
)

type action struct {
	typ     actionType
	version *library.Version
}

func askPath(suggested string) string {
	prompt := promptui.Prompt{
		Label:   "Edit Destination",
		Default: suggested,
	}
	name, _ := prompt.Run()
	return name
}

func actionsOnDocument(l library.Library, d library.Document) {
	items := []string{"🔙 Back"}
	var actions []action

	if d.LocalPath != "" {
		items = append(items, "open locally")
		actions = append(actions, action{openLocally, nil})
		items = append(items, "open local folder")
		actions = append(actions, action{openFolder, nil})
		items = append(items, "delete")
		actions = append(actions, action{deletelocal, nil})
		if d.State == library.Modified || d.State == library.Conflict {
			items = append(items, "send update to the pool")
			actions = append(actions, action{uploadLocal, nil})
		}
	}

	for _, v := range d.Versions {
		i, _ := security.IdentityFromId(v.AuthorId)
		switch {
		case v.State == library.Updated:
			items = append(items, fmt.Sprintf("receive update from %s: size %d, date %s", i.Nick, v.Size, v.ModTime.Format(time.Stamp)))
			actions = append(actions, action{updateLocal, &v})
		case v.State == library.Conflict:
			items = append(items, fmt.Sprintf("receive replacement from %s: size %d, date %s", i.Nick, v.Size, v.ModTime.Format(time.Stamp)))
			actions = append(actions, action{updateLocal, &v})
		}
		items = append(items, fmt.Sprintf("download from %s: size %d, date %s", i.Nick, v.Size, v.ModTime.Format(time.Stamp)))
		actions = append(actions, action{downloadTemp, &v})
	}

	label := fmt.Sprintf("Choose the action on '%s'", d.Name)
	prompt := promptui.Select{
		Label: label,
		Items: items,
	}
	idx, _, _ := prompt.Run()
	if idx == 0 {
		return
	}
	a := actions[idx-1]
	switch a.typ {
	case openLocally:
		open.Start(d.LocalPath)
	case openFolder:
		open.Start(filepath.Dir(d.LocalPath))
	case deletelocal:
		os.Remove(d.LocalPath)
	case uploadLocal:
		l.Send(d.LocalPath, d.Name, true)
	case updateLocal:
		localPath := d.LocalPath
		if localPath == "" {
			localPath = askPath(filepath.Join(xdg.UserDirs.Documents, l.Pool.Name, d.Name))
		}
		_, err := l.Receive(a.version.Id, localPath)
		if err == nil {
			color.Green("File updated to %s", localPath)
		} else {
			color.Red("Cannot update to %s: %v", localPath, err)
		}
	case downloadTemp:
		dest := askPath(filepath.Join(os.TempDir(), d.Name))
		err := l.Save(a.version.Id, dest)
		if err == nil {
			color.Green("File downloaded to %s", dest)
		} else {
			color.Red("Cannot download to %s: %v", dest, err)
		}
	}

}

func documentFormat(d library.Document) string {
	var icon, author string
	switch d.State {
	case library.Sync:
		icon = "✓"
	case library.New:
		icon = "⇐"
	case library.Updated:
		icon = "←"
	case library.Modified:
		icon = "→"
	case library.Conflict:
		icon = "⇆"
	case library.Deleted:
		icon = "🗑"
	}

	if identity, ok, _ := security.GetIdentity(d.AuthorId); ok {
		author = identity.Nick
	} else {
		author = d.AuthorId
	}
	return fmt.Sprintf("%s %s 👤%s 🔗%s", icon, d.Name, author, d.LocalPath)
}

func Library(p *pool.Pool) {

	l := library.Get(p, "library")

	folder := ""
	for {
		ls, err := l.List(folder)
		if core.IsErr(err, "cannot read document list: %v") {
			color.Red("something wrong")
			return
		}

		items := []string{"🔙 Back", "⟳ Refresh", "＋ Add"}
		for _, s := range ls.Subfolders {
			items = append(items, fmt.Sprintf("📁 %s", s))
		}
		for _, d := range ls.Documents {
			items = append(items, documentFormat(d))
		}

		prompt := promptui.Select{
			Label: "Choose",
			Items: items,
		}

		idx, _, err := prompt.Run()
		if err != nil {
			return
		}
		switch idx {
		case 0:
			if folder == "" {
				return
			} else {
				folder = path.Dir(folder)
				folder = strings.TrimLeft(folder, ".")
			}
		case 1:
			p.Sync()
		case 2:
			addDocument(l)
		default:
			if idx < len(ls.Subfolders)+3 {
				folder = ls.Subfolders[idx-3]
			} else {
				actionsOnDocument(l, ls.Documents[idx-3-len(ls.Subfolders)])
			}
		}
	}
}
