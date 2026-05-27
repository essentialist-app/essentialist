# Agent Instructions

## Project Purpose

This project is flashcard app composed of a GUI (`cmd/essentialist`) and a command-line version (`cmd/flashdown`).

## File Structure

The project is structured as follows:
- **`cmd/`**: Contains the main entry points for the applications.
- **`internal/`**: Contains the core business logic and internal components of the application.
- **`third_party/`**: Contains external dependencies, specifically related to LaTeX/TeX processing.
- **`internal/samples/testdata/`**: Contains sample data files used for testing.
- **`README.md`**: Contains general project information.
- **`LICENSE`**: Contains the project license.
- **`.github/`**: Contains GitHub Actions workflows for CI/CD.

```
├── essentialist
│   ├── about.go
│   ├── congrats.go
│   ├── error.go
│   ├── file.go
│   ├── flatpak
│   │   ├── go.mod.yml
│   │   ├── io.github.essentialist_app.essentialist.desktop
│   │   ├── io.github.essentialist_app.essentialist.metainfo.xml
│   │   ├── io.github.essentialist_app.essentialist.png
│   │   ├── io.github.essentialist_app.essentialist-rocket.png
│   │   ├── io.github.essentialist_app.essentialist-rocket.svg
│   │   ├── io.github.essentialist_app.essentialist.svg
│   │   ├── io.github.essentialist_app.essentialist.yml
│   │   └── modules.txt
│   ├── FyneApp.toml
│   ├── help.go
│   ├── home.go
│   ├── i18n
│   │   ├── i18n.go
│   │   ├── i18n_test.go
│   │   └── locales
│   │       ├── de.json
│   │       ├── en.json
│   │       ├── es.json
│   │       ├── fr.json
│   │       ├── it.json
│   │       ├── ja.json
│   │       ├── ko.json
│   │       ├── zh.json
│   │       └── zh-TW.json
│   ├── licenses.md
│   ├── licenses.tpl
│   ├── main.go
│   ├── math.go
│   ├── math_parser.go
│   ├── math_parser_test.go
│   ├── math_renderer.go
│   ├── math_test.go
│   ├── question.go
│   ├── screen.go
│   ├── settings.go
│   ├── splash.go
│   ├── table.go
│   └── utils.go
└── flashdown
    ├── main.go
    └── markdown_area.go
internal
├── accessor.go
├── card.go
├── card_test.go
├── deck.go
├── deck_test.go
├── game.go
├── markdown.go
├── markdown_test.go
├── meta.go
├── meta_test.go
└── samples
    ├── sample.md
    ├── test
    │   ├── emoji.md
    │   ├── icon.png
    │   ├── image.md
    │   ├── list.md
    │   ├── maths.md
    │   └── table.md
    └── testdata
        ├── test-1.md
        └── test-2.md
```

## Coding Style
The codebase is be written in Go, following standard Go conventions.

## Features

The project include functionality for:
- Mathematical parsing and rendering (`cmd/essentialist/math_parser.go`, `cmd/essentialist/math.go`).
- Card/Deck management (`internal/card.go`, `internal/deck.go`).
- Internationalization (i18n) support for multiple languages (`cmd/essentialist/i18n/locales/`).
- User interface/display components (`cmd/essentialist/screen.go`, `cmd/essentialist/splash.go`).
- LaTeX/TeX related processing, suggesting a focus on document generation or complex formatting.cmd
