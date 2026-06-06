package main

import (
	_ "embed"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/essentialist-app/essentialist/cmd/essentialist/i18n"
)

//go:embed licenses.md
var licenses string

type AboutScreen struct {
}

func NewAboutScreen() Screen {
	return &AboutScreen{}
}

func (s *AboutScreen) Show(app Application) {
	back := widget.NewButton("Settings", func() {
		app.Display(NewSettingsScreen())
	})
	topBar := newTopBar(i18n.MustLocalize("about"), back)
	app.SetContent(
		container.NewBorder(
			topBar,
			nil,
			nil,
			nil,
			topBar,
			container.NewVScroll(
				container.NewVBox(
					widget.NewLabel("Essentialist"),
					NewRichTextFromMarkdown("Site: [https://essentialist.app](https://essentialist.app)"),
					NewRichTextFromMarkdown("Author: [github.com/lugu](https://github.com/lugu)"),
					widget.NewLabel("Version: 0.4.0"),
					NewRichTextFromMarkdown(licenses),
				),
			),
		),
	)
	app.SetOnTypedKey(EscapeKeyHandler(app, NewSettingsScreen()))
}

func (s *AboutScreen) Hide(app Application) {
}
