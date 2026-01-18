package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	cp "github.com/marver003/razpravljalnica/api/controlplane"
	"github.com/marver003/razpravljalnica/internal/server"
	"github.com/rivo/tview"
)

type Dashboard struct {
	app        *tview.Application
	cp         *server.ControlPlane
	chainView  *tview.TextView
	nodeTable  *tview.Table
	logView    *tview.TextView
	msgChannel chan string
}

func Start(cp *server.ControlPlane) {
	d := &Dashboard{
		app:        tview.NewApplication(),
		cp:         cp,
		msgChannel: make(chan string, 100),
	}

	// setup logging
	cp.SetLogger(func(format string, v ...interface{}) {
		msg := fmt.Sprintf(format, v...)
		timestamp := time.Now().Format("15:04:05")
		d.msgChannel <- fmt.Sprintf("[%s] %s", timestamp, msg)
	})

	// setup UI components
	d.setupLayout()

	// start update loops
	go d.logLoop()
	go d.refreshLoop()

	if err := d.app.Run(); err != nil {
		panic(err)
	}
}

func (d *Dashboard) setupLayout() {
	// Chain Visualization (Top)
	d.chainView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	d.chainView.SetBorder(true).SetTitle(" Active Replication Chain ")

	// Node Details (Middle)
	d.nodeTable = tview.NewTable().
		SetBorders(true).
		SetSelectable(false, false)
	d.nodeTable.SetTitle(" Registered Nodes ").SetBorder(true)

	d.setupTableHeader()

	// Logs (Bottom)
	d.logView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetMaxLines(1000)
	d.logView.SetTitle(" Live Logs (Tab to focus/scroll) ").SetBorder(true)

	// Help Footer
	help := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Tab: Switch Focus | PgUp/PgDn: Scroll Logs (when focused) | Esc: Quit")

	// Layout: Chain (20%), Table (50%), Logs (30%), Help (Fixed)
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(d.chainView, 6, 1, false).
		AddItem(d.nodeTable, 0, 2, true). // default focus on table
		AddItem(d.logView, 0, 1, false).
		AddItem(help, 1, 0, false)

	// enable tab to switch focus between table and logs
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			if d.app.GetFocus() == d.nodeTable {
				d.app.SetFocus(d.logView)
			} else {
				d.app.SetFocus(d.nodeTable)
			}
			return nil
		}
		if event.Key() == tcell.KeyEsc {
			d.app.Stop()
			return nil
		}
		return event
	})

	d.app.SetRoot(flex, true)
}

func (d *Dashboard) setupTableHeader() {
	headers := []string{"ID", "Address", "Role", "Last Seen", "Status"}
	for i, h := range headers {
		d.nodeTable.SetCell(0, i,
			tview.NewTableCell(h).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignCenter).
				SetAttributes(tcell.AttrBold))
	}
}

func (d *Dashboard) logLoop() {
	for msg := range d.msgChannel {
		// Capture msg in closure
		message := msg
		d.app.QueueUpdateDraw(func() {
			fmt.Fprintf(d.logView, "%s\n", message)
			d.logView.ScrollToEnd()
		})
	}
}

func (d *Dashboard) refreshLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	for range ticker.C {
		chain, nodes := d.cp.GetState()

		d.app.QueueUpdateDraw(func() {
			d.updateChainView(chain)
			d.updateNodeTable(chain, nodes)
		})
	}
}

func (d *Dashboard) updateChainView(chain []*cp.NodeInfo) {
	if len(chain) == 0 {
		d.chainView.SetText("[red]Empty Chain[-]")
		return
	}

	var sb strings.Builder
	for i, node := range chain {
		roleColor := "white"
		roleStr := ""

		isHead := (i == 0)
		isTail := (i == len(chain)-1)

		if isHead && isTail {
			roleColor = "yellow"
			roleStr = "(HEAD/TAIL)"
		} else if isHead {
			roleColor = "green"
			roleStr = "(HEAD)"
		} else if isTail {
			roleColor = "blue"
			roleStr = "(TAIL)"
		}

		// node representation
		fmt.Fprintf(&sb, "[%s]%s %s[-]", roleColor, node.NodeId, roleStr)

		// arrow
		if i < len(chain)-1 {
			sb.WriteString(" [yellow]➔[-] ")
		}
	}
	// add a newline for vertical centering in the 6-line high box
	d.chainView.SetText("\n" + sb.String())
}

func (d *Dashboard) updateNodeTable(chain []*cp.NodeInfo, nodes map[string]*server.NodeStatus) {
	// clear existing rows (keep header)
	rowCount := d.nodeTable.GetRowCount()
	if rowCount > 1 {
		for r := rowCount - 1; r > 0; r-- {
			d.nodeTable.RemoveRow(r)
		}
	}

	// helper to check chain role
	getRole := func(id string) string {
		if len(chain) == 0 {
			return "-"
		}

		isHead := chain[0].NodeId == id
		isTail := chain[len(chain)-1].NodeId == id

		if isHead && isTail {
			return "HEAD/TAIL"
		}
		if isHead {
			return "HEAD"
		}
		if isTail {
			return "TAIL"
		}

		for _, n := range chain {
			if n.NodeId == id {
				return "Replica"
			}
		}
		return "-"
	}

	// sort nodes by ID
	ids := make([]string, 0, len(nodes))
	for id := range nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for i, id := range ids {
		status := nodes[id]
		row := i + 1

		role := getRole(id)

		lastSeenDur := time.Since(status.LastSeen).Round(time.Second)
		statusStr := "[green]Active[-]"
		if lastSeenDur > 5*time.Second {
			statusStr = "[red]Timeout[-]"
		}

		d.nodeTable.SetCell(row, 0, tview.NewTableCell(id).SetAlign(tview.AlignCenter))
		d.nodeTable.SetCell(row, 1, tview.NewTableCell(status.Info.Address).SetAlign(tview.AlignCenter))

		// colorize role
		roleColor := tcell.ColorWhite
		if role == "HEAD/TAIL" {
			roleColor = tcell.ColorYellow
		}
		if role == "HEAD" {
			roleColor = tcell.ColorGreen
		}
		if role == "TAIL" {
			roleColor = tcell.ColorBlue
		}
		d.nodeTable.SetCell(row, 2, tview.NewTableCell(role).SetTextColor(roleColor).SetAlign(tview.AlignCenter))

		d.nodeTable.SetCell(row, 3, tview.NewTableCell(lastSeenDur.String()).SetAlign(tview.AlignCenter))
		d.nodeTable.SetCell(row, 4, tview.NewTableCell(statusStr).SetAlign(tview.AlignCenter))
	}
}
