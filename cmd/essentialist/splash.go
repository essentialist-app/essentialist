package main

import (
	"runtime"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/essentialist-app/essentialist/cmd/essentialist/i18n"
)

var (
	welcomeMessage = ""
)

type SplashScreen struct{}

func (s *SplashScreen) load(app Application) {
	decks, err := loadDecks()
	if err != nil {
		app.Display(NewErrorScreen(err))
		return
	}
	sort.SliceStable(decks, func(i, j int) bool {
		return decks[i].DeckName() < decks[j].DeckName()
	})
	app.Display(NewHomeScreen(decks))
}

func newWelcomeTopBar(app Application) *fyne.Container {
	home := widget.NewButton(i18n.MustLocalize("settings"), func() {
		app.Display(NewSettingsScreen())
	})
	return newTopBar(i18n.MustLocalize("welcome"), home)
}

func (s *SplashScreen) keyHandler(app Application) func(*fyne.KeyEvent) {
	return func(key *fyne.KeyEvent) {
		if key.Name != "" {
			switch key.Name {
			case fyne.KeyQ, fyne.KeyEscape:
				app.Close()
			case fyne.KeyReturn, fyne.KeySpace:
				app.Display(NewSettingsScreen())
			}
		} else {
			switch key.Physical {
			case fyne.HardwareKey{ScanCode: 9}, fyne.HardwareKey{ScanCode: 24}: // Escape
				app.Close()
			case fyne.HardwareKey{ScanCode: 36}, fyne.HardwareKey{ScanCode: 55}: // Enter or Space
				app.Display(NewSettingsScreen())
			}
		}
	}
}

func (s *SplashScreen) Show(app Application) {
	welcomeMessage = i18n.MustLocalize("please_set_directory")
	// Welcome message when the application is launched for the first time.
	prefs := fyne.CurrentApp().Preferences()
	dir := prefs.StringWithFallback("directory", "")
	if dir == "" && runtime.GOOS != "js" && runtime.GOOS != "wasm" {
		topBar := newWelcomeTopBar(app)
		welcomeButton := widget.NewButton(welcomeMessage, func() {
			app.Display(NewSettingsScreen())
		})
		center := container.NewVScroll(welcomeButton)
		app.SetContent(container.New(layout.NewBorderLayout(
			topBar, nil, nil, nil), topBar, center))
		app.SetOnTypedKey(s.keyHandler(app))
		return
	}
	emptyContainer := container.New(layout.NewHBoxLayout(), layout.NewSpacer())
	app.SetContent(emptyContainer)
	go s.load(app) // load the games in the background
}

func (s *SplashScreen) Hide(app Application) {}

func NewSplashScreen() Screen {
	return &SplashScreen{}
}
