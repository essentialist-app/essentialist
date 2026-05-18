package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/essentialist-app/essentialist/cmd/essentialist/i18n"
	"github.com/essentialist-app/essentialist/internal"
)

type maxWidthCenterLayout struct {
	width float32
}

// NewCenterWithMaxWidthLayout returns a layout which centers its content
// elements given them no more than width.
func NewMaxWidthCenterLayout(width float32) fyne.Layout {
	return &maxWidthCenterLayout{width}
}

func (c *maxWidthCenterLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	pos := fyne.NewPos(0, 0)
	if size.Width > c.width {
		pos.X = (size.Width - c.width) / 2
		size.Width = c.width
	}
	for _, child := range objects {
		child.Resize(size)
		child.Move(pos)
	}
}

func (c *maxWidthCenterLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minSize := fyne.NewSize(0, 0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		minSize = minSize.Max(child.MinSize())
	}
	return minSize
}

func space() fyne.CanvasObject {
	return layout.NewSpacer()
}

type QuestionScreen struct {
	game *internal.Game
}

func NewQuestionScreen(game *internal.Game) Screen {
	return &QuestionScreen{game: game}
}

func (s *QuestionScreen) keyHandler(app Application) func(*fyne.KeyEvent) {
	return func(key *fyne.KeyEvent) {
		if key.Name != "" {
			switch key.Name {
			case fyne.KeySpace, fyne.KeyReturn:
				app.Display(NewAnswerScreen(s.game))
			case fyne.KeyQ, fyne.KeyEscape:
				s.game.Save()
				app.Display(NewSplashScreen())
			case fyne.KeyS, fyne.KeyN:
				s.game.Skip()
				if s.game.IsFinished() {
					app.Display(NewCongratsScreen(s.game))
				} else {
					app.Display(NewQuestionScreen(s.game))
				}
			case fyne.KeyP:
				s.game.Previous()
				app.Display(NewQuestionScreen(s.game))
			case fyne.KeyW:
				s.game.Save()
			}
		} else {
			switch key.Physical {
			case fyne.HardwareKey{ScanCode: 9}, fyne.HardwareKey{ScanCode: 24}: // Escape
				s.game.Save()
				app.Display(NewSplashScreen())
			case fyne.HardwareKey{ScanCode: 39}, fyne.HardwareKey{ScanCode: 57}: // 's' or 'n'
				s.game.Skip()
				if s.game.IsFinished() {
					app.Display(NewCongratsScreen(s.game))
				} else {
					app.Display(NewQuestionScreen(s.game))
				}
			case fyne.HardwareKey{ScanCode: 33}: // 'p'
				s.game.Previous()
				app.Display(NewQuestionScreen(s.game))
			case fyne.HardwareKey{ScanCode: 25}: // 'w'
				s.game.Save()
			}
		}
	}
}

// TODO: make the test selectable
func card(md string, path string) fyne.CanvasObject {
	richText := NewRichTextFromMarkdownAt(md, path)
	width := richText.MinSize().Width
	richText.Wrapping = fyne.TextWrapWord
	return container.New(NewMaxWidthCenterLayout(width), richText)
}

func (s *QuestionScreen) Show(app Application) {
	topBar := newProgressTopBar(app, s.game)
	currentCard := s.game.CurrentCard()
	question := card("### "+currentCard.Question, currentCard.Deck.Path)
	button := continueButton(app, s.game)

	vbox := container.New(layout.NewVBoxLayout(), topBar, space(), question,
		space(), button)
	app.SetContent(vbox)
	app.SetOnTypedKey(s.keyHandler(app))
}

func (s *QuestionScreen) Hide(app Application) {
	app.SetOnTypedKey(nil)
}

type AnswerScreen struct {
	game *internal.Game
}

func NewAnswerScreen(game *internal.Game) Screen {
	return &AnswerScreen{game: game}
}

func (s *AnswerScreen) answersButton(app Application) *fyne.Container {
	bt := func(label string, score internal.Score) *widget.Button {
		return widget.NewButton(label,
			func() {
				s.reviewScore(app, score)
			})
	}
	buttons := []fyne.CanvasObject{
		bt(i18n.MustLocalize("score_again"), internal.ScoreAgain),
		bt(i18n.MustLocalize("score_hard"), internal.ScoreHard),
		bt(i18n.MustLocalize("score_good"), internal.ScoreGood),
		bt(i18n.MustLocalize("score_easy"), internal.ScoreEasy),
	}
	return container.New(layout.NewGridLayout(4), buttons...)
}

func (s *AnswerScreen) reviewScore(app Application, score internal.Score) {
	s.game.Review(score)
	if s.game.IsFinished() {
		s.game.Save()
		app.Display(NewCongratsScreen(s.game))
	} else {
		app.Display(NewQuestionScreen(s.game))
	}
}

func (s *AnswerScreen) keyHandler(app Application) func(*fyne.KeyEvent) {
	return func(key *fyne.KeyEvent) {
		if key.Name != "" {
			switch key.Name {
			case fyne.Key1:
				s.reviewScore(app, internal.ScoreAgain)
			case fyne.Key2:
				s.reviewScore(app, internal.ScoreHard)
			case fyne.Key3:
				s.reviewScore(app, internal.ScoreGood)
			case fyne.Key4:
				s.reviewScore(app, internal.ScoreEasy)
			case fyne.KeyQ, fyne.KeyEscape:
				s.game.Save()
				app.Display(NewSplashScreen())
			case fyne.KeyS, fyne.KeyN:
				s.game.Skip()
				if s.game.IsFinished() {
					app.Display(NewCongratsScreen(s.game))
				} else {
					app.Display(NewQuestionScreen(s.game))
				}
			case fyne.KeyP:
				s.game.Previous()
				app.Display(NewQuestionScreen(s.game))
			case fyne.KeyW:
				s.game.Save()
			}
		} else {
			switch key.Physical {
			case fyne.HardwareKey{ScanCode: 9}, fyne.HardwareKey{ScanCode: 24}: // Escape
				s.game.Save()
				app.Display(NewSplashScreen())
			case fyne.HardwareKey{ScanCode: 39}, fyne.HardwareKey{ScanCode: 57}: // 's' or 'n'
				s.game.Skip()
				if s.game.IsFinished() {
					app.Display(NewCongratsScreen(s.game))
				} else {
					app.Display(NewQuestionScreen(s.game))
				}
			case fyne.HardwareKey{ScanCode: 33}: // 'p'
				s.game.Previous()
				app.Display(NewQuestionScreen(s.game))
			case fyne.HardwareKey{ScanCode: 25}: // 'w'
				s.game.Save()
			}
		}
	}
}

func (s *AnswerScreen) Show(app Application) {
	topBar := newProgressTopBar(app, s.game)

	currentCard := s.game.CurrentCard()
	question := card("### "+currentCard.Question, currentCard.Deck.Path)
	line := canvas.NewLine(color.Gray16{0xaaaa})
	answer := card(currentCard.Answer, currentCard.Deck.Path)

	buttons := s.answersButton(app)
	vbox := container.New(layout.NewVBoxLayout(), topBar, space(), question,
		space(), line, space(), answer, space(), buttons)
	app.SetContent(vbox)
	app.SetOnTypedKey(s.keyHandler(app))
}

func (s *AnswerScreen) Hide(app Application) {
	app.SetOnTypedKey(nil)
}