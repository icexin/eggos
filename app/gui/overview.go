package gui

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"time"

	"github.com/aarzilli/nucular"
	ncommand "github.com/aarzilli/nucular/command"
	"github.com/aarzilli/nucular/label"
	"github.com/aarzilli/nucular/rect"
	nstyle "github.com/aarzilli/nucular/style"

	"golang.org/x/mobile/event/mouse"
)

type OptionEnum int

const (
	OptionA = OptionEnum(iota)
	OptionB
	OptionC
)

type overviewDemo struct {
	ShowMenu    bool
	Titlebar    bool
	Border      bool
	Resize      bool
	Movable     bool
	NoScrollbar bool
	Minimizable bool
	Close       bool

	HeaderAlign nstyle.HeaderAlign

	// Menu status
	Mprog   int
	Mslider int
	Mcheck  bool
	Prog    int
	Slider  int
	Check   bool

	// Basic widgets status
	IntSlider                          int
	FloatSlider                        float64
	ProgValue                          int
	PropertyFloat                      float64
	PropertyInt                        int
	PropertyNeg                        int
	RangeMin, RangeValue, RangeMax     float64
	RangeIntMin, RangeInt, RangeIntMax int
	Checkbox                           bool
	Option                             OptionEnum

	// Selectable
	Selected  []bool
	Selected2 []bool

	// Combo
	CurrentWeapon              int
	CheckValues                []bool
	Position                   []float64
	ComboColor                 color.RGBA
	ProgA, ProgB, ProgC, ProgD int
	Weapons                    []string
	ColMode                    int
	TimeSelected               int
	Text0Editor, Text1Editor   nucular.TextEditor
	FieldEditor                nucular.TextEditor
	BoxEditor                  nucular.TextEditor

	// Popup
	PSelect []bool
	PProg   int
	PSlider int

	// Layout
	GroupTitlebar           bool
	GroupBorder             bool
	GroupNoScrollbar        bool
	GroupWidth, GroupHeight int

	GroupSelected []bool

	// Vertical Split
	A, B, C    int
	HA, HB, HC int

	Img *image.RGBA

	Resizing1, Resizing2, Resizing3, Resizing4 bool

	edEntry1, edEntry2, edEntry3 nucular.TextEditor

	Theme nstyle.Theme

	FitWidthEditor nucular.TextEditor
	FitWidthIdx    int
}

func newOverviewDemo() (od *overviewDemo) {
	od = &overviewDemo{}
	od.ShowMenu = true
	od.Titlebar = true
	od.Border = true
	od.Resize = true
	od.Movable = true
	od.NoScrollbar = false
	od.Close = true
	od.HeaderAlign = nstyle.HeaderRight
	od.Mprog = 60
	od.Mslider = 8
	od.Mcheck = true
	od.Prog = 40
	od.Slider = 10
	od.Check = true
	od.IntSlider = 5
	od.FloatSlider = 2.5
	od.ProgValue = 40
	od.Option = OptionA
	od.Selected = []bool{false, false, true, false}
	od.Selected2 = []bool{true, false, false, false, false, true, false, false, false, false, true, false, false, false, false, true}
	od.CurrentWeapon = 0
	od.CheckValues = make([]bool, 5)
	od.Position = make([]float64, 3)
	od.ComboColor = color.RGBA{130, 50, 50, 255}
	od.ProgA = 20
	od.ProgB = 40
	od.ProgC = 10
	od.ProgD = 90
	od.Weapons = []string{"Fist", "Pistol", "Shotgun", "Plasma", "BFG"}
	od.PSelect = make([]bool, 4)
	od.PProg = 0
	od.PSlider = 10

	//Layout
	od.GroupTitlebar = false
	od.GroupBorder = true
	od.GroupNoScrollbar = false

	od.GroupSelected = make([]bool, 16)

	od.A = 100
	od.B = 100
	od.C = 100

	od.HA = 100
	od.HB = 100
	od.HC = 100

	od.PropertyFloat = 2
	od.PropertyInt = 10
	od.PropertyNeg = 10

	od.RangeMin = 0
	od.RangeValue = 50
	od.RangeMax = 100

	od.RangeIntMin = 0
	od.RangeInt = 2048
	od.RangeIntMax = 4096

	od.GroupWidth = 320
	od.GroupHeight = 200

	fh, err := os.Open("rob.pike.mixtape.jpg")
	if err == nil {
		defer fh.Close()
		img, _ := jpeg.Decode(fh)
		od.Img = image.NewRGBA(img.Bounds())
		draw.Draw(od.Img, img.Bounds(), img, image.Point{}, draw.Src)
	}

	od.Text0Editor.Flags = nucular.EditSimple
	od.Text0Editor.Maxlen = 64
	od.FieldEditor.Flags = nucular.EditField
	od.FieldEditor.Maxlen = 64
	od.BoxEditor.Flags = nucular.EditBox | nucular.EditNeverInsertMode
	od.Text1Editor.Flags = nucular.EditField | nucular.EditSigEnter
	od.Text1Editor.Maxlen = 64

	od.edEntry1.Flags = nucular.EditSimple
	od.edEntry1.Buffer = []rune("Menu Item 1")
	od.edEntry2.Flags = nucular.EditSimple
	od.edEntry2.Buffer = []rune("Menu Item 2")
	od.edEntry3.Flags = nucular.EditSimple
	od.edEntry3.Buffer = []rune("Menu Item 3")

	od.FitWidthEditor.Flags = nucular.EditSimple
	od.FitWidthEditor.Buffer = []rune("test")

	return od
}

func (od *overviewDemo) overviewDemo(w *nucular.Window) {
	keybindings(w)
	mw := w.Master()

	style := mw.Style()
	style.NormalWindow.Header.Align = od.HeaderAlign

	if od.ShowMenu {
		od.overviewMenubar(w)
	}

	if w.TreePush(nucular.TreeTab, "Window", false) {
		w.Row(30).Dynamic(2)
		w.CheckboxText("Titlebar", &od.Titlebar)
		w.CheckboxText("Menu", &od.ShowMenu)
		w.CheckboxText("Border", &od.Border)
		w.CheckboxText("Resizable", &od.Resize)
		w.CheckboxText("Movable", &od.Movable)
		w.CheckboxText("No Scrollbars", &od.NoScrollbar)
		w.CheckboxText("Closable", &od.Close)
		w.TreePop()
	}

	if w.TreePush(nucular.TreeTab, "Widgets", false) {
		w.Row(20).Dynamic(1)
		if w.TreePush(nucular.TreeNode, "Text", false) {
			od.overviewTextWidgets(w)
			w.TreePop()
		}

		w.Row(20).Dynamic(1)
		if w.TreePush(nucular.TreeNode, "Button", false) {
			w.Row(30).Static(100, 100, 100)
			if w.Button(label.T("Button"), false) {
				fmt.Printf("Button pressed!\n")
			}
			if w.Button(label.T("Repeater"), true) {
				fmt.Printf("Repeater is being pressed\n")
			}
			w.Button(label.C(color.RGBA{0x00, 0x00, 0xff, 0xff}), false)
			w.TreePop()
		}

		w.Row(20).Dynamic(1)
		if w.TreePush(nucular.TreeNode, "Basic", false) {
			od.overviewBasicWidgets(w)
			w.TreePop()
		}

		w.Row(20).Dynamic(1)
		if w.TreePush(nucular.TreeNode, "Selectable", false) {
			od.overviewSelectableWidgets(w)
			w.TreePop()
		}

		w.Row(20).Dynamic(1)
		if w.TreePush(nucular.TreeNode, "Combo", false) {
			od.overviewComboWidgets(w)
			w.TreePop()
		}

		w.Row(20).Dynamic(1)
		if w.TreePush(nucular.TreeNode, "Input", false) {
			w.Row(25).Static(120, 150)
			w.Label("Default:", "LC")
			od.Text0Editor.Edit(w)
			w.Label("Field:", "LC")
			od.FieldEditor.Edit(w)
			w.Label("Box:", "LC")
			w.Row(180).Static(278)
			od.BoxEditor.Edit(w)
			w.Row(25).Static(120, 150)
			active := od.Text1Editor.Edit(w)
			if w.Button(label.T("Submit"), false) || (active&nucular.EditCommitted != 0) {
				od.BoxEditor.Buffer = append(od.BoxEditor.Buffer, od.Text1Editor.Buffer...)
				od.BoxEditor.Buffer = append(od.BoxEditor.Buffer, '\n')
				od.Text1Editor.Buffer = []rune{}
			}
			w.TreePop()
		}

		w.TreePop()
	}

	if w.TreePush(nucular.TreeTab, "Popup", false) {
		od.overviewPopup(w)
		w.TreePop()
	}

	if w.TreePush(nucular.TreeTab, "Layout", false) {
		od.overviewLayout(w)
		w.TreePop()
	}

	if w.TreePush(nucular.TreeTab, "Image & Custom", false) {
		if od.Img != nil {
			w.RowScaled(335).StaticScaled(500)
			w.Image(od.Img)
		} else {
			w.Row(25).Dynamic(1)
			w.Label("could not load example image", "LC")
		}

		w.RowScaled(335).StaticScaled(500)
		bounds, out := w.Custom(nstyle.WidgetStateInactive)
		if out != nil {
			style := mw.Style()
			od.drawCustomWidget(bounds, style, out)
		}

		w.TreePop()
	}
}

func (od *overviewDemo) overviewMenubar(w *nucular.Window) {
	w.MenubarBegin()
	w.Row(25).Static(45, 45, 70, 70, 70)
	if w := w.Menu(label.TA("MENU", "CC"), 120, nil); w != nil {
		w.Row(25).Dynamic(1)
		if w.MenuItem(label.TA("Hide", "LC")) {
			od.ShowMenu = false
		}
		if w.MenuItem(label.TA("About", "LC")) {
			od.showAppAbout(w.Master())
		}
		w.Progress(&od.Prog, 100, true)
		w.SliderInt(0, &od.Slider, 16, 1)
		w.CheckboxText("check", &od.Check)
		perf := w.Master().GetPerf()
		w.CheckboxText("Show perf", &perf)
		w.Master().SetPerf(perf)
		if w.MenuItem(label.TA("Close", "LC")) {
			go w.Master().Close()
		}
	}
	if w := w.Menu(label.TA("THEME", "CC"), 180, nil); w != nil {
		w.Row(25).Dynamic(1)
		newtheme := od.Theme
		if w.OptionText("Default Theme", newtheme == nstyle.DefaultTheme) {
			newtheme = nstyle.DefaultTheme
		}
		if w.OptionText("White Theme", newtheme == nstyle.WhiteTheme) {
			newtheme = nstyle.WhiteTheme
		}
		if w.OptionText("Red Theme", newtheme == nstyle.RedTheme) {
			newtheme = nstyle.RedTheme
		}
		if w.OptionText("Dark Theme", newtheme == nstyle.DarkTheme) {
			newtheme = nstyle.DarkTheme
		}
		if newtheme != od.Theme {
			od.Theme = newtheme
			w.Master().SetStyle(nstyle.FromTheme(od.Theme, w.Master().Style().Scaling))
			w.Close()
		}
	}
	w.Progress(&od.Mprog, 100, true)
	w.SliderInt(0, &od.Mslider, 16, 1)
	w.CheckboxText("check", &od.Mcheck)
	w.MenubarEnd()
}

func (od *overviewDemo) overviewTextWidgets(w *nucular.Window) {
	w.Row(20).Dynamic(1)
	w.Label("Label aligned left", "LC")
	w.Label("Label aligned centered", "CC")
	w.Label("Label aligned right", "RC")
	w.LabelColored("Blue text", "LC", color.RGBA{0x00, 0x00, 0xff, 0xff})
	w.LabelColored("Yellow text", "LC", color.RGBA{0xff, 0xff, 0x00, 0xff})
	w.Row(100).Static(200)
	w.LabelWrap("This is a very long line to hopefully get this text to be wrapped into multiple lines to show line wrapping")
	w.Row(100).Dynamic(1)
	w.LabelWrap("This is another long text to show dynamic window changes on multiline text")
}

func (od *overviewDemo) overviewBasicWidgets(w *nucular.Window) {
	w.Row(30).Static(100)
	w.CheckboxText("Checkbox", &od.Checkbox)

	w.Row(30).Static(80, 80, 80)
	if w.OptionText("optionA", od.Option == OptionA) {
		od.Option = OptionA
	}
	if w.OptionText("optionB", od.Option == OptionB) {
		od.Option = OptionB
	}
	if w.OptionText("optionC", od.Option == OptionC) {
		od.Option = OptionC
	}

	w.Row(30).Static(170, 200)
	w.Label("Slider int", "LC")
	w.SliderInt(0, &od.IntSlider, 10, 1)

	w.Label("Slider float", "LC")
	w.SliderFloat(0, &od.FloatSlider, 5.0, 0.5)
	w.Label(fmt.Sprintf("Progressbar %d", od.ProgValue), "LC")
	w.Progress(&od.ProgValue, 100, true)

	w.Label("Property float:", "LC")
	w.PropertyFloat("Float:", 0, &od.PropertyFloat, 64.0, 0.1, 0.2, 2)
	w.Label("Property int:", "LC")
	w.PropertyInt("Int:", 0, &od.PropertyInt, 100, 1, 1)
	w.Label("Property neg:", "LC")
	w.PropertyInt("Neg:", -10, &od.PropertyNeg, 10, 1, 1)

	w.Row(25).Dynamic(1)
	w.Label("Range:", "LC")
	w.Row(25).Dynamic(3)

	w.PropertyFloat("#min:", 0, &od.RangeMin, od.RangeMax, 1.0, 0.2, 3)
	w.PropertyFloat("#float:", od.RangeMin, &od.RangeValue, od.RangeMax, 1.0, 0.2, 3)
	w.PropertyFloat("#max:", od.RangeMin, &od.RangeMax, 100, 1.0, 0.2, 3)

	w.PropertyInt("#min:", -100, &od.RangeIntMin, od.RangeIntMax, 1, 10)
	w.PropertyInt("#neg:", od.RangeIntMin, &od.RangeInt, od.RangeIntMax, 1, 10)
	w.PropertyInt("#max:", od.RangeIntMin, &od.RangeIntMax, 100, 1, 10)
}

func (od *overviewDemo) overviewSelectableWidgets(w *nucular.Window) {
	w.Row(20).Dynamic(1)
	if w.TreePush(nucular.TreeNode, "List", false) {
		w.Row(18).Static(130)
		w.SelectableLabel("Selectable", "LC", &od.Selected[0])
		w.SelectableLabel("Selectable", "LC", &od.Selected[1])
		w.Label("Not Selectable", "LC")
		w.SelectableLabel("Selectable", "LC", &od.Selected[2])
		w.SelectableLabel("Selectable", "LC", &od.Selected[3])
		w.TreePop()
	}
	w.Row(20).Dynamic(1)
	if w.TreePush(nucular.TreeNode, "Grid", false) {
		w.Row(50).Static(50, 50, 50, 50)
		for i := 0; i < len(od.Selected2); i++ {
			if w.SelectableLabel("Z", "CC", &od.Selected2[i]) {
				x := (i % 4)
				y := i / 4
				if x > 0 {
					od.Selected2[i-1] = !od.Selected2[i-1]
				}
				if x < 3 {
					od.Selected2[i+1] = !od.Selected2[i+1]
				}
				if y > 0 {
					od.Selected2[i-4] = !od.Selected2[i-4]
				}
				if y < 3 {
					od.Selected2[i+4] = !od.Selected2[i+4]
				}
			}
		}
		w.TreePop()
	}
}

func (od *overviewDemo) overviewComboWidgets(w *nucular.Window) {
	w.Row(25).Static(200)

	// Default combo box
	od.CurrentWeapon = w.ComboSimple(od.Weapons, od.CurrentWeapon, 25)

	// Slider color combobox
	if w := w.Combo(label.C(od.ComboColor), 200, nil); w != nil {
		w.Row(30).Ratio(0.15, 0.85)

		slider := func(lbl string, b *uint8) {
			w.Label(lbl, "LC")
			i := int(*b)
			w.SliderInt(0, &i, 255, 5)
			*b = uint8(i)
		}

		slider("R:", &od.ComboColor.R)
		slider("G:", &od.ComboColor.G)
		slider("B:", &od.ComboColor.B)
	}

	// Progressbar combobox
	sum := od.ProgA + od.ProgB + od.ProgC + od.ProgD
	if w := w.Combo(label.T(fmt.Sprintf("%d", sum)), 200, nil); w != nil {
		w.Row(30).Dynamic(1)
		w.Progress(&od.ProgA, 100, true)
		w.Progress(&od.ProgB, 100, true)
		w.Progress(&od.ProgC, 100, true)
		w.Progress(&od.ProgD, 100, true)
	}

	// Checkbox combobox
	sum = 0
	for _, v := range od.CheckValues {
		if v {
			sum++
		}
	}
	if w := w.Combo(label.T(fmt.Sprintf("%d", sum)), 200, nil); w != nil {
		w.Row(30).Dynamic(1)
		w.CheckboxText(od.Weapons[0], &od.CheckValues[0])
		w.CheckboxText(od.Weapons[1], &od.CheckValues[1])
		w.CheckboxText(od.Weapons[2], &od.CheckValues[2])
		w.CheckboxText(od.Weapons[3], &od.CheckValues[3])
	}

	// Complex text combobox
	if w := w.Combo(label.T(fmt.Sprintf("%.2f %.2f %.2f", od.Position[0], od.Position[1], od.Position[2])), 200, nil); w != nil {
		w.Row(25).Dynamic(1)
		w.PropertyFloat("#X:", -1024.0, &od.Position[0], 1024.0, 1, 0.5, 2)
		w.PropertyFloat("#Y:", -1024.0, &od.Position[1], 1024.0, 1, 0.5, 2)
		w.PropertyFloat("#X:", -1024.0, &od.Position[2], 1024.0, 1, 0.5, 2)
	}

	// Time combobox
	var selt time.Time
	if od.TimeSelected == 0 {
		selt = time.Now()
	} else {
		selt = time.Unix(int64(od.TimeSelected), 0)
	}

	if w := w.Combo(label.T(fmt.Sprintf("%02d:%02d:%02d", selt.Hour(), selt.Minute(), selt.Second())), 250, nil); w != nil {
		var selt time.Time
		if od.TimeSelected == 0 {
			selt = time.Now()
		} else {
			selt = time.Unix(int64(od.TimeSelected), 0)
		}

		w.Row(25).Dynamic(1)
		var second, minute, hour int = selt.Second(), selt.Minute(), selt.Hour()
		fmt.Printf("time selected: %v\n", selt)
		w.PropertyInt("#S:", 0, &second, 60, 1, 1)
		w.PropertyInt("#M:", 0, &minute, 60, 1, 1)
		w.PropertyInt("#H:", 0, &hour, 60, 1, 1)
		od.TimeSelected = int(time.Date(selt.Year(), selt.Month(), selt.Day(), hour, minute, second, 0, selt.Location()).Unix())
	}
}

func (od *overviewDemo) errorPopup(w *nucular.Window) {
	w.Row(25).Dynamic(1)
	w.Label("A terrible error has occured", "LC")
	w.Row(25).Dynamic(2)
	if w.Button(label.T("OK"), false) {
		w.Close()
	}
	if w.Button(label.T("Cancel"), false) {
		w.Close()
	}
}

func (od *overviewDemo) overviewPopup(w *nucular.Window) {
	// Menu contextual
	w.Row(30).Dynamic(1)
	w.Label("Right click me for menu", "LC")
	if w := w.ContextualOpen(0, image.Point{100, 300}, w.LastWidgetBounds, nil); w != nil {
		w.Row(25).Dynamic(1)
		w.CheckboxText("Menu", &od.ShowMenu)
		w.Progress(&od.PProg, 100, true)
		w.SliderInt(0, &od.Slider, 16, 1)
		if w.MenuItem(label.TA("About", "CC")) {
			od.showAppAbout(w.Master())
		}
		sel := func(i int) string {
			if od.PSelect[i] {
				return "Unselect"
			}
			return "Select"
		}
		w.SelectableLabel(sel(0), "LC", &od.PSelect[0])
		w.SelectableLabel(sel(1), "LC", &od.PSelect[1])
		w.SelectableLabel(sel(2), "LC", &od.PSelect[2])
		w.SelectableLabel(sel(3), "LC", &od.PSelect[3])
	}

	w.Label("Right click me for a simple autoresizing menu", "LC")
	if w := w.ContextualOpen(0, image.Point{}, w.LastWidgetBounds, nil); w != nil {
		w.Row(25).Dynamic(1)
		w.MenuItem(label.TA(string(od.edEntry1.Buffer), "LC"))
		w.MenuItem(label.TA(string(od.edEntry2.Buffer), "LC"))
		w.MenuItem(label.TA(string(od.edEntry3.Buffer), "LC"))
	}

	w.Row(30).Static(100, 150)
	w.Label("Menu Item 1:", "LC")
	od.edEntry1.Edit(w)
	w.Label("Menu Item 2:", "LC")
	od.edEntry2.Edit(w)
	w.Label("Menu Item 3:", "LC")
	od.edEntry3.Edit(w)

	// Popup
	w.Row(30).Static(100, 50)
	w.Label("Popup:", "LC")
	if w.Button(label.T("popup"), false) {
		w.Master().PopupOpen("Error", nucular.WindowMovable|nucular.WindowTitle|nucular.WindowDynamic|nucular.WindowNoScrollbar, rect.Rect{20, 100, 230, 150}, true, od.errorPopup)
	}

	// Tooltip
	w.Row(30).Static(180)
	w.Label("Hover me for tooltip", "LC")
	if w.Input().Mouse.HoveringRect(w.LastWidgetBounds) {
		w.Tooltip("This is a tooltip")
	}

	// Second tooltip
	w.Row(30).Static(420)
	w.Label("Can also hover me for a different tooltip", "LC")
	if w.Input().Mouse.HoveringRect(w.LastWidgetBounds) {
		w.Tooltip("This is another tooltip")
	}
}

func (od *overviewDemo) overviewLayout(w *nucular.Window) {
	w.Row(20).Dynamic(1)
	if w.TreePush(nucular.TreeNode, "Widget", false) {
		btn := func() {
			w.Button(label.T("button"), false)
		}

		w.Row(30).Dynamic(1)
		w.Label("Dynamic fixed column layout with generated position and size (LayoutRowDynamic):", "LC")
		w.Row(30).Dynamic(3)
		btn()
		btn()
		btn()

		w.Row(30).Dynamic(1)
		w.Label("static fixed column layout with generated position and size (LayoutRowStatic):", "LC")
		w.Row(30).Static(100, 100, 100)
		btn()
		btn()
		btn()

		w.Row(30).Dynamic(1)
		w.Label("Dynamic array-based custom column layout with generated position and custom size (LayoutRowRatio):", "LC")
		w.Row(30).Ratio(0.2, 0.6, 0.2)
		btn()
		btn()
		btn()

		w.Row(30).Dynamic(1)
		w.Label("Static custom column layout with generated position and custom size (LayoutRowStatic + LayoutSetWidth):", "LC")
		w.Row(30).Static()
		w.LayoutSetWidth(100)
		btn()
		w.LayoutSetWidth(200)
		btn()
		w.LayoutSetWidth(50)
		btn()

		w.Row(30).Dynamic(1)
		w.Label("Static custom column layout with generated position and automatically fit size (LayoutRowStatic + LayoutFitWidth)", "LC")
		w.Row(30).Static(100, 100, 100)
		w.Label("Content:", "LC")
		od.FitWidthEditor.Edit(w)
		if w.ButtonText("Recalculate") {
			od.FitWidthIdx++
		}
		for i := 0; i < 3; i++ {
			w.Row(30).Static()
			w.LayoutFitWidth(od.FitWidthIdx, 1)
			w.Label(fmt.Sprintf("%02d:", i), "LC")
			w.LayoutFitWidth(od.FitWidthIdx, 100)
			w.Label(string(od.FitWidthEditor.Buffer), "LC")
			w.LayoutFitWidth(od.FitWidthIdx, 1)
			w.Label("END", "LC")
		}

		w.Row(30).Dynamic(1)
		w.Label("Static array-based custom column layout with dynamic space in the middle (LayoutRowStatic + LayoutResetStatic):", "LC")
		w.Row(30).Static(100, 100)
		btn()
		btn()
		w.LayoutResetStatic(0, 100, 100)
		w.Spacing(1)
		btn()
		btn()

		w.Row(30).Dynamic(1)
		w.Label("Dynamic immediate mode custom column layout with generated position and custom size (LayoutRowRatio):", "LC")
		w.Row(30).Ratio(0.2, 0.6, 0.2)
		btn()
		btn()
		btn()

		w.Row(30).Dynamic(1)
		w.Label("Static immediate mode custom column layout with generated position and custom size (LayoutRowStatic):", "LC")
		w.Row(30).Static(100, 200, 50)
		btn()
		btn()
		btn()

		w.Row(30).Dynamic(1)
		w.Label("Static free space with custom position and custom size (LayoutSpaceBegin + LayoutSpacePush):", "LC")
		w.Row(120).SpaceBegin(4)
		w.LayoutSpacePush(rect.Rect{100, 0, 100, 30})
		btn()
		w.LayoutSpacePush(rect.Rect{0, 15, 100, 30})
		btn()
		w.LayoutSpacePush(rect.Rect{200, 15, 100, 30})
		btn()
		w.LayoutSpacePush(rect.Rect{100, 30, 100, 30})
		btn()

		w.TreePop()
	}

	w.Row(20).Dynamic(1)
	if w.TreePush(nucular.TreeNode, "Group", false) {
		groupFlags := nucular.WindowFlags(0)
		if od.GroupBorder {
			groupFlags |= nucular.WindowBorder
		}
		if od.GroupNoScrollbar {
			groupFlags |= nucular.WindowNoScrollbar
		}
		if od.GroupTitlebar {
			groupFlags |= nucular.WindowTitle
		}
		groupFlags |= nucular.WindowNoHScrollbar

		w.Row(30).Dynamic(3)
		w.CheckboxText("Titlebar", &od.GroupTitlebar)
		w.CheckboxText("Border", &od.GroupBorder)
		w.CheckboxText("No Scrollbar", &od.GroupNoScrollbar)

		w.Row(22).Static(50, 130, 130)
		w.Label("size:", "LC")
		w.PropertyInt("#Width:", 100, &od.GroupWidth, 500, 10, 10)
		w.PropertyInt("#Height:", 100, &od.GroupHeight, 500, 10, 10)

		w.Row(od.GroupHeight).Static(od.GroupWidth, od.GroupWidth)
		if sw := w.GroupBegin("Group", groupFlags); sw != nil {
			sw.Row(18).Static(100)
			for i := range od.GroupSelected {
				sel := "Unselected"
				if od.GroupSelected[i] {
					sel = "Selected"
				}
				sw.SelectableLabel(sel, "CC", &od.GroupSelected[i])
			}
			sw.GroupEnd()
		}
		w.TreePop()
	}

	w.Row(20).Dynamic(1)
	if w.TreePush(nucular.TreeNode, "Simple", false) {
		w.Row(300).Dynamic(2)
		if sw := w.GroupBegin("Group Without Border", 0); sw != nil {
			sw.Row(18).Static(150)
			for i := 0; i < 64; i++ {
				sw.Label(fmt.Sprintf("%#02x", i), "LC")
			}
			sw.GroupEnd()
		}

		if sw := w.GroupBegin("Group With Border", 0); sw != nil {
			sw.Row(25).Dynamic(2)
			for i := 0; i < 64; i++ {
				sw.Button(label.T(fmt.Sprintf("%08d", (((i%7)*10)^32)+(64+(i%2)*2))), false)
			}
			sw.GroupEnd()
		}
		w.TreePop()
	}

	w.Row(20).Dynamic(1)
	if w.TreePush(nucular.TreeNode, "Complex", false) {
		w.Row(500).SpaceBegin(64)
		w.LayoutSpacePush(rect.Rect{0, 0, 150, 500})
		if sw := w.GroupBegin("Group Left", nucular.WindowBorder); sw != nil {
			sw.Row(18).Static(100)
			for i := range od.GroupSelected {
				txt := "Unselected"
				if od.GroupSelected[i] {
					txt = "Selected"
				}
				sw.SelectableLabel(txt, "CC", &od.GroupSelected[i])
			}
			sw.GroupEnd()
		}

		w.LayoutSpacePush(rect.Rect{160, 0, 150, 240})
		if sw := w.GroupBegin("Group top", nucular.WindowBorder); sw != nil {
			sw.Row(25).Dynamic(1)
			sw.Button(label.T("#FFAA"), false)
			sw.Button(label.T("#FFBB"), false)
			sw.Button(label.T("#FFCC"), false)
			sw.Button(label.T("#FFDD"), false)
			sw.Button(label.T("#FFEE"), false)
			sw.Button(label.T("#FFFF"), false)
			sw.GroupEnd()
		}

		w.LayoutSpacePush(rect.Rect{160, 250, 150, 250})
		if sw := w.GroupBegin("Group bottom", nucular.WindowBorder); sw != nil {
			sw.Row(25).Dynamic(1)
			sw.Button(label.T("#FFAA"), false)
			sw.Button(label.T("#FFBB"), false)
			sw.Button(label.T("#FFCC"), false)
			sw.Button(label.T("#FFDD"), false)
			sw.Button(label.T("#FFEE"), false)
			sw.Button(label.T("#FFFF"), false)
			sw.GroupEnd()
		}

		w.LayoutSpacePush(rect.Rect{320, 0, 150, 150})
		if sw := w.GroupBegin("Group right top", nucular.WindowBorder); sw != nil {
			sw.Row(18).Static(100)
			for i := range od.GroupSelected {
				txt := "Unselected"
				if od.GroupSelected[i] {
					txt = "Selected"
				}
				sw.SelectableLabel(txt, "CC", &od.GroupSelected[i])
			}
			sw.GroupEnd()
		}

		w.LayoutSpacePush(rect.Rect{320, 160, 150, 150})
		if sw := w.GroupBegin("Group right center", nucular.WindowBorder); sw != nil {
			sw.Row(18).Static(100)
			for i := range od.GroupSelected {
				txt := "Unselected"
				if od.GroupSelected[i] {
					txt = "Selected"
				}
				sw.SelectableLabel(txt, "CC", &od.GroupSelected[i])
			}
			sw.GroupEnd()
		}

		w.LayoutSpacePush(rect.Rect{320, 320, 150, 150})
		if sw := w.GroupBegin("Group right bottom", nucular.WindowBorder); sw != nil {
			sw.Row(18).Static(100)
			for i := range od.GroupSelected {
				txt := "Unselected"
				if od.GroupSelected[i] {
					txt = "Selected"
				}
				sw.SelectableLabel(txt, "CC", &od.GroupSelected[i])
			}
			sw.GroupEnd()
		}

		w.TreePop()
	}

	w.Row(20).Dynamic(1)
	if w.TreePush(nucular.TreeNode, "Splitter", false) {
		w.Row(20).Static(320)
		w.Label("Use slider and spinner to change tile size", "LC")
		w.Label("Drag the space between tiles to change tile ratio", "LC")

		w.Row(20).Dynamic(1)
		if w.TreePush(nucular.TreeNode, "Vertical", false) {
			w.Row(30).Static(100, 100)
			w.Label("left:", "LC")
			w.SliderInt(10, &od.A, 200, 10)
			w.Label("middle:", "LC")
			w.SliderInt(10, &od.B, 200, 10)
			w.Label("right:", "LC")
			w.SliderInt(10, &od.C, 200, 10)

			w.Row(200).Static(od.A, 8, od.B, 8, od.C)

			// Left column
			if sw := w.GroupBegin("left", nucular.WindowNoScrollbar|nucular.WindowBorder); sw != nil {
				sw.Row(25).Dynamic(1)
				sw.Button(label.T("#FFAA"), false)
				sw.Button(label.T("#FFBB"), false)
				sw.Button(label.T("#FFCC"), false)
				sw.Button(label.T("#FFDD"), false)
				sw.Button(label.T("#FFEE"), false)
				sw.Button(label.T("#FFFF"), false)
				sw.GroupEnd()
			}

			// Scaler (a custom widget)
			if w.CustomState() == nstyle.WidgetStateActive {
				od.Resizing1 = true
			}
			if od.Resizing1 {
				if !w.Input().Mouse.Down(mouse.ButtonLeft) {
					od.Resizing1 = false
				} else {
					od.A += w.Input().Mouse.Delta.X
					od.B -= w.Input().Mouse.Delta.X
				}
			}
			w.Custom(nstyle.WidgetStateInactive)

			// Middle column
			if sw := w.GroupBegin("center", nucular.WindowNoScrollbar|nucular.WindowBorder); sw != nil {
				sw.Row(25).Dynamic(1)
				sw.Button(label.T("#FFAA"), false)
				sw.Button(label.T("#FFBB"), false)
				sw.Button(label.T("#FFCC"), false)
				sw.Button(label.T("#FFDD"), false)
				sw.Button(label.T("#FFEE"), false)
				sw.Button(label.T("#FFFF"), false)
				sw.GroupEnd()
			}

			// Scaler (a custom widget)
			if w.CustomState() == nstyle.WidgetStateActive {
				od.Resizing2 = true
			}
			if od.Resizing2 {
				if !w.Input().Mouse.Down(mouse.ButtonLeft) {
					od.Resizing2 = false
				} else {
					od.B += w.Input().Mouse.Delta.X
					od.C -= w.Input().Mouse.Delta.X
				}
			}
			w.Custom(nstyle.WidgetStateInactive)

			if sw := w.GroupBegin("right", nucular.WindowNoScrollbar|nucular.WindowBorder); sw != nil {
				sw.Row(25).Dynamic(1)
				sw.Button(label.T("#FFAA"), false)
				sw.Button(label.T("#FFBB"), false)
				sw.Button(label.T("#FFCC"), false)
				sw.Button(label.T("#FFDD"), false)
				sw.Button(label.T("#FFEE"), false)
				sw.Button(label.T("#FFFF"), false)
				sw.GroupEnd()
			}

			w.TreePop()
		}

		w.Row(20).Dynamic(1)
		if w.TreePush(nucular.TreeNode, "Horizontal", false) {
			w.Row(30).Static(100, 100)
			w.Label("top:", "LC")
			w.SliderInt(10, &od.HA, 200, 10)
			w.Label("middle:", "LC")
			w.SliderInt(10, &od.HB, 200, 10)
			w.Label("bottom:", "LC")
			w.SliderInt(10, &od.HC, 200, 10)

			w.Row(od.HA).Dynamic(1)
			if sw := w.GroupBegin("top", nucular.WindowNoScrollbar|nucular.WindowBorder); sw != nil {
				sw.Row(25).Dynamic(3)
				sw.Button(label.T("#FFAA"), false)
				sw.Button(label.T("#FFBB"), false)
				sw.Button(label.T("#FFCC"), false)
				sw.Button(label.T("#FFDD"), false)
				sw.Button(label.T("#FFEE"), false)
				sw.Button(label.T("#FFFF"), false)
				sw.GroupEnd()
			}

			// Scaler (a custom widget)
			w.Row(8).Dynamic(1)
			if w.CustomState() == nstyle.WidgetStateActive {
				od.Resizing3 = true
			}
			if od.Resizing3 {
				if !w.Input().Mouse.Down(mouse.ButtonLeft) {
					od.Resizing3 = false
				} else {
					od.HA += w.Input().Mouse.Delta.Y
					od.HB -= w.Input().Mouse.Delta.Y
				}
			}
			w.Custom(nstyle.WidgetStateInactive)

			w.Row(od.HB).Dynamic(1)
			if sw := w.GroupBegin("middle", nucular.WindowNoScrollbar|nucular.WindowBorder); sw != nil {
				sw.Row(25).Dynamic(3)
				sw.Button(label.T("#FFAA"), false)
				sw.Button(label.T("#FFBB"), false)
				sw.Button(label.T("#FFCC"), false)
				sw.Button(label.T("#FFDD"), false)
				sw.Button(label.T("#FFEE"), false)
				sw.Button(label.T("#FFFF"), false)
				sw.GroupEnd()
			}

			// Scaler (a custom widget)
			w.Row(8).Dynamic(1)
			if w.CustomState() == nstyle.WidgetStateActive {
				od.Resizing4 = true
			}
			if od.Resizing4 {
				if !w.Input().Mouse.Down(mouse.ButtonLeft) {
					od.Resizing4 = false
				} else {
					od.HB += w.Input().Mouse.Delta.Y
					od.HC -= w.Input().Mouse.Delta.Y
				}
			}
			w.Custom(nstyle.WidgetStateInactive)

			w.Row(od.HC).Dynamic(1)
			if sw := w.GroupBegin("right", nucular.WindowNoScrollbar|nucular.WindowBorder); sw != nil {
				sw.Row(25).Dynamic(3)
				sw.Button(label.T("#FFAA"), false)
				sw.Button(label.T("#FFBB"), false)
				sw.Button(label.T("#FFCC"), false)
				sw.Button(label.T("#FFDD"), false)
				sw.Button(label.T("#FFEE"), false)
				sw.Button(label.T("#FFFF"), false)
				sw.GroupEnd()
			}

			w.TreePop()
		}
		w.TreePop()
	}
}

func (od *overviewDemo) drawCustomWidget(bounds rect.Rect, style *nstyle.Style, out *ncommand.Buffer) {
	clr := color.RGBA{0xff, 0x00, 0xff, 0xff}
	bl := image.Point{bounds.X, bounds.Y + bounds.H - 1}
	tr := image.Point{bounds.X + bounds.W - 1, bounds.Y}
	br := image.Point{bounds.X + bounds.W - 1, bounds.Y + bounds.H - 1}
	out.StrokeLine(bounds.Min(), bl, 1, clr)
	out.StrokeLine(bl, br, 1, clr)
	out.StrokeLine(br, tr, 1, clr)
	out.StrokeLine(tr, bounds.Min(), 1, clr)
	out.StrokeLine(bounds.Min().Add(image.Point{50, 50}), bounds.Max().Add(image.Point{-50, -50}), 5, clr)
}

func (od *overviewDemo) aboutPopup(w *nucular.Window) {
	w.Row(20).Dynamic(1)
	w.Label("Nucular", "LC")
	w.Label("By Alessandro Arzilli", "LC")
	w.Label("based on Nuklear by Micha Mettke", "LC")
}

func (od *overviewDemo) showAppAbout(mw nucular.MasterWindow) {
	var wf nucular.WindowFlags

	if od.Border {
		wf |= nucular.WindowBorder
	}
	if od.Resize {
		wf |= nucular.WindowScalable
	}
	if od.Movable {
		wf |= nucular.WindowMovable
	}
	if od.NoScrollbar {
		wf |= nucular.WindowNoScrollbar
	}
	if od.Close {
		wf |= nucular.WindowClosable
	}
	if od.Titlebar {
		wf |= nucular.WindowTitle
	}

	mw.PopupOpen("About", wf, rect.Rect{20, 100, 300, 190}, true, od.aboutPopup)
}
