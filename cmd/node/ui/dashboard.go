package ui

import (
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/marver003/razpravljalnica/internal/server"
	"github.com/rivo/tview"
)

type logWriter struct {
	msgChannel chan string
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.msgChannel <- string(p)
	return len(p), nil
}

type Dashboard struct {
	app        *tview.Application
	node       *server.Node
	headerView *tview.TextView
	logView    *tview.TextView
	msgChannel chan string
}

func Start(node *server.Node) {
	ch := make(chan string, 500)
	d := &Dashboard{
		app:        tview.NewApplication(),
		node:       node,
		msgChannel: ch,
	}

	// redirect global log to this channel
	log.SetOutput(&logWriter{msgChannel: ch})

	// setup UI
	d.setupLayout()

	// start update loops
	go d.logLoop()
	go d.refreshLoop()

	if err := d.app.Run(); err != nil {
		panic(err)
	}
}

func (d *Dashboard) setupLayout() {
	// Header (Node Identity)
	d.headerView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	d.headerView.SetBorder(true).SetTitle(" Node Identity ")

	// Logs (Bottom)
	d.logView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetMaxLines(1000)
	d.logView.SetTitle(" Node Logs (Scrollable) ").SetBorder(true)

	// Main Layout
	help := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("Esc: Quit")

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(d.headerView, 3, 1, false).
		AddItem(d.logView, 0, 1, true). // take all remaining space
		AddItem(help, 1, 0, false)

	// input capture
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			d.app.Stop()
			return nil
		}
		return event
	})

	d.app.SetRoot(flex, true)
}

func (d *Dashboard) logLoop() {
	for msg := range d.msgChannel {
		message := msg
		d.app.QueueUpdateDraw(func() {
			fmt.Fprintf(d.logView, "%s", message)
			d.logView.ScrollToEnd()
		})
	}
}

func (d *Dashboard) refreshLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	for range ticker.C {
		state := d.node.GetUIState()
		d.app.QueueUpdateDraw(func() {
			d.updateHeader(state)
		})
	}
}

func (d *Dashboard) updateHeader(state server.NodeUIState) {
	role := "REPLICA"
	color := "white"
	if state.IsHead && state.IsTail {
		role = "HEAD/TAIL"
		color = "yellow"
	} else if state.IsHead {
		role = "HEAD"
		color = "green"
	} else if state.IsTail {
		role = "TAIL"
		color = "blue"
	}

	d.headerView.SetText(fmt.Sprintf("ID: [::b]%s[::-] | Address: [::b]%s[::-] | Role: [%s::b]%s[-]",
		state.ID, state.Address, color, role))
}
