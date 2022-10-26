package gui

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"time"

	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/font"
	"github.com/aarzilli/nucular/label"
	"github.com/aarzilli/nucular/rect"
	nstyle "github.com/aarzilli/nucular/style"
	styled "github.com/aarzilli/nucular/style-editor"
	"github.com/jspc/eggos/app"
	"github.com/jspc/eggos/app/shiny"

	"golang.org/x/mobile/event/key"
)

const dotrace = false
const scaling = 1

//var theme nucular.Theme = nucular.WhiteTheme
var theme nstyle.Theme = nstyle.DarkTheme

func id(fn func(*nucular.Window)) func() func(*nucular.Window) {
	return func() func(*nucular.Window) {
		return fn
	}
}

type Demo struct {
	Name     string
	Title    string
	Flags    nucular.WindowFlags
	UpdateFn func() func(*nucular.Window)
}

var demos = []Demo{
	{"button", "Button Demo", 0, id(buttonDemo)},
	{"basic", "Basic Demo", 0, func() func(*nucular.Window) {
		go func() {
			for {
				time.Sleep(1 * time.Second)
				if Wnd.Closed() {
					break
				}
				Wnd.Changed()
			}
		}()
		return basicDemo
	}},
	{"basic2", "Text Editor Demo", 0, textEditorDemo},
	{"calc", "Calculator Demo", 0, func() func(*nucular.Window) {
		var cd calcDemo
		cd.current = &cd.a
		return cd.calculatorDemo
	}},
	{"overview", "Overview", 0, func() func(*nucular.Window) {
		od := newOverviewDemo()
		od.Theme = theme
		return od.overviewDemo
	}},
	{"editor", "Multiline Text Editor", nucular.WindowNoScrollbar, multilineTextEditorDemo},
	{"split", "Split panel demo", nucular.WindowNoScrollbar, func() func(*nucular.Window) {
		pd := &panelDebug{}
		pd.Init()
		return pd.Update
	}},
	{"nestedmenu", "Nested menu demo", 0, id(nestedMenu)},
	{"list", "List", nucular.WindowNoScrollbar, id(listDemo)},
	{"styled", "Theme Editor", 0, func() func(*nucular.Window) {
		return styled.StyleEditor(nil, func(out string) { fmt.Println(out) })
	}},
}

var Wnd nucular.MasterWindow

func main(ctx *app.Context) error {
	nucular.DriverMain = shiny.Main

	whichdemo := ""
	if len(ctx.Args) > 1 {
		whichdemo = ctx.Args[1]
	}

	switch whichdemo {
	case "multi", "":
		Wnd = nucular.NewMasterWindow(0, "Multiwindow Demo", func(w *nucular.Window) {})
		Wnd.PopupOpen("Multiwindow Demo", nucular.WindowTitle|nucular.WindowBorder|nucular.WindowMovable|nucular.WindowScalable|nucular.WindowNonmodal, rect.Rect{0, 0, 400, 300}, true, multiDemo)
	default:
		for i := range demos {
			if demos[i].Name == whichdemo {
				Wnd = nucular.NewMasterWindow(demos[i].Flags, demos[i].Title, demos[i].UpdateFn())
				break
			}
		}
	}
	if Wnd == nil {
		fmt.Fprintf(os.Stderr, "unknown demo %q\n", whichdemo)
		fmt.Fprintf(os.Stderr, "known demos:\n")
		for i := range demos {
			fmt.Fprintf(os.Stderr, "\t%s\n", demos[i].Name)
		}
		return nil
	}

	Wnd.SetStyle(nstyle.FromTheme(theme, scaling))

	normalFontData, normerr := ioutil.ReadFile("demofont.ttf")
	if normerr == nil {
		szf := 12 * scaling
		face, normerr := font.NewFace(normalFontData, int(szf))
		if normerr == nil {
			style := Wnd.Style()
			style.Font = face
		}
	}

	Wnd.Main()
	return nil
}

func buttonDemo(w *nucular.Window) {
	w.Row(20).Static(60, 60)
	if w.Button(label.T("button1"), false) {
		fmt.Printf("button pressed!\n")
	}
	if w.Button(label.T("button2"), false) {
		fmt.Printf("button 2 pressed!\n")
	}
}

type difficulty int

const (
	easy = difficulty(iota)
	hard
)

var op difficulty = easy
var compression int

func basicDemo(w *nucular.Window) {
	w.Row(30).Dynamic(1)
	w.Label(time.Now().Format("15:04:05"), "RT")

	w.Row(30).Static(80)
	if w.Button(label.T("button"), false) {
		fmt.Printf("button pressed! difficulty: %v compression: %d\n", op, compression)
	}
	w.Row(30).Dynamic(2)
	if w.OptionText("easy", op == easy) {
		op = easy
	}
	if w.OptionText("hard", op == hard) {
		op = hard
	}
	w.Row(25).Dynamic(1)
	w.PropertyInt("Compression:", 0, &compression, 100, 10, 1)
}

func textEditorDemo() func(w *nucular.Window) {
	var textEditorEditor nucular.TextEditor
	textEditorEditor.Flags = nucular.EditSelectable
	textEditorEditor.Buffer = []rune("prova")
	return func(w *nucular.Window) {
		w.Row(30).Dynamic(1)
		textEditorEditor.Maxlen = 30
		textEditorEditor.Edit(w)
	}
}

func multilineTextEditorDemo() func(w *nucular.Window) {
	var multilineTextEditor nucular.TextEditor
	bs, _ := ioutil.ReadFile("overview.go")
	multilineTextEditor.Buffer = []rune(string(bs))
	var ibeam bool
	return func(w *nucular.Window) {
		w.Row(20).Dynamic(1)
		w.CheckboxText("I-Beam cursor", &ibeam)
		w.Row(0).Dynamic(1)
		multilineTextEditor.Flags = nucular.EditMultiline | nucular.EditSelectable | nucular.EditClipboard
		if ibeam {
			multilineTextEditor.Flags |= nucular.EditIbeamCursor
		}
		multilineTextEditor.Edit(w)
	}
}

type panelDebug struct {
	splitv          nucular.ScalableSplit
	splith          nucular.ScalableSplit
	showblocks      bool
	showsingleblock bool
	showtabs        bool
}

func (pd *panelDebug) Init() {
	pd.splitv.MinSize = 80
	pd.splitv.Size = 120
	pd.splitv.Spacing = 5
	pd.splith.MinSize = 100
	pd.splith.Size = 300
	pd.splith.Spacing = 5
	pd.showtabs = true
	pd.showsingleblock = true
}

func (pd *panelDebug) Update(w *nucular.Window) {
	for _, k := range w.Input().Keyboard.Keys {
		if k.Rune == 'b' {
			pd.showsingleblock = false
			pd.showblocks = !pd.showblocks
		}
		if k.Rune == 'B' {
			pd.showsingleblock = !pd.showsingleblock
		}
		if k.Rune == 't' {
			pd.showtabs = !pd.showtabs
		}
	}

	if pd.showtabs {
		w.Row(20).Dynamic(2)
		w.Label("A", "LC")
		w.Label("B", "LC")
	}

	area := w.Row(0).SpaceBegin(0)

	if pd.showsingleblock {
		w.LayoutSpacePushScaled(area)
		bounds, out := w.Custom(nstyle.WidgetStateInactive)
		if out != nil {
			out.FillRect(bounds, 10, color.RGBA{0x00, 0x00, 0xff, 0xff})
		}
	} else {
		leftbounds, rightbounds := pd.splitv.Vertical(w, area)
		viewbounds, commitbounds := pd.splith.Horizontal(w, rightbounds)

		w.LayoutSpacePushScaled(leftbounds)
		pd.groupOrBlock(w, "index-files", nucular.WindowBorder)

		w.LayoutSpacePushScaled(viewbounds)
		pd.groupOrBlock(w, "index-diff", nucular.WindowBorder)

		w.LayoutSpacePushScaled(commitbounds)
		pd.groupOrBlock(w, "index-right-column", nucular.WindowNoScrollbar|nucular.WindowBorder)
	}
}

func (pd *panelDebug) groupOrBlock(w *nucular.Window, name string, flags nucular.WindowFlags) {
	if pd.showblocks {
		bounds, out := w.Custom(nstyle.WidgetStateInactive)
		if out != nil {
			out.FillRect(bounds, 10, color.RGBA{0x00, 0x00, 0xff, 0xff})
		}
	} else {
		if sw := w.GroupBegin(name, flags); sw != nil {
			sw.GroupEnd()
		}
	}
}

func nestedMenu(w *nucular.Window) {
	w.Row(20).Static(180)
	w.Label("Test", "CC")
	w.ContextualOpen(0, image.Point{0, 0}, w.LastWidgetBounds, func(w *nucular.Window) {
		w.Row(20).Dynamic(1)
		if w.MenuItem(label.TA("Submenu", "CC")) {
			w.ContextualOpen(0, image.Point{0, 0}, rect.Rect{0, 0, 0, 0}, func(w *nucular.Window) {
				w.Row(20).Dynamic(1)
				if w.MenuItem(label.TA("Done", "CC")) {
					fmt.Printf("done\n")
				}
			})
		}
	})
}

var listDemoSelected = -1
var listDemoCnt = 0

func listDemo(w *nucular.Window) {
	const N = 100
	recenter := false
	for _, e := range w.Input().Keyboard.Keys {
		switch e.Code {
		case key.CodeDownArrow:
			listDemoSelected++
			if listDemoSelected >= N {
				listDemoSelected = N - 1
			}
			recenter = true
		case key.CodeUpArrow:
			listDemoSelected--
			if listDemoSelected < -1 {
				listDemoSelected = -1
			}
			recenter = true
		}
	}
	w.Row(0).Dynamic(1)
	if gl, w := nucular.GroupListStart(w, N, "list", nucular.WindowNoHScrollbar); w != nil {
		if !recenter {
			gl.SkipToVisible(20)
		}
		w.Row(20).Dynamic(1)
		cnt := 0
		for gl.Next() {
			cnt++
			i := gl.Index()
			selected := i == listDemoSelected
			w.SelectableLabel(fmt.Sprintf("label %d", i), "LC", &selected)
			if selected {
				listDemoSelected = i
				if recenter {
					gl.Center()
				}
			}
		}
		if cnt != listDemoCnt {
			listDemoCnt = cnt
			fmt.Printf("called %d times\n", listDemoCnt)
		}
	}
}

func keybindings(w *nucular.Window) {
	mw := w.Master()
	if in := w.Input(); in != nil {
		k := in.Keyboard
		for _, e := range k.Keys {
			scaling := mw.Style().Scaling
			switch {
			case (e.Modifiers == key.ModControl || e.Modifiers == key.ModControl|key.ModShift) && (e.Code == key.CodeEqualSign):
				mw.Style().Scale(scaling + 0.1)
			case (e.Modifiers == key.ModControl || e.Modifiers == key.ModControl|key.ModShift) && (e.Code == key.CodeHyphenMinus):
				mw.Style().Scale(scaling - 0.1)
			case (e.Modifiers == key.ModControl) && (e.Code == key.CodeF):
				mw.SetPerf(!mw.GetPerf())
			}
		}
	}
}

func multiDemo(w *nucular.Window) {
	keybindings(w)

	w.Row(20).Dynamic(1)
	w.Label("Welcome to the multi-window demo.", "LC")
	w.Label("Open any demo window by clicking on the buttons.", "LC")
	w.Label("To run a demo as a stand-alone window use:", "LC")
	w.Label("     \"./uidemo <demo-name>\"", "LC")
	w.Row(30).Static(100, 100, 100)
	for i := range demos {
		if w.ButtonText(demos[i].Name) {
			w.Master().PopupOpen(demos[i].Title, nucular.WindowDefaultFlags|nucular.WindowNonmodal|demos[i].Flags, rect.Rect{0, 0, 200, 200}, true, demos[i].UpdateFn())
		}
	}
}

func init() {
	app.Register("uidemo", main)
}
