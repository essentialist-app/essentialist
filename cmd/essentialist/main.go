package main

import (
	_ "embed"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/essentialist-app/essentialist/cmd/essentialist/i18n"
)

//go:embed licenses.md
var iconBytes []byte

func parseArgs() {
	if len(os.Args) == 2 {
		overrideDirectory = os.Args[1]
	}
}

func main() {
	parseArgs()
	application := app.NewWithID("io.github.essentialist_app.essentialist")
	lang := application.Preferences().StringWithFallback("language", "en")
	i18n.SetLanguage(lang)


	application.Settings().SetTheme(getTheme())
	window := application.NewWindow("Essentialist")
	icon := fyne.NewStaticResource("io.github.essentialist_app.essentialist.png", iconBytes)
	application.SetIcon(icon)
	window.Resize(fyne.NewSize(640, 480))
	NewApplication(application, window).Display(NewSplashScreen())
	window.ShowAndRun()
}
