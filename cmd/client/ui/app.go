package ui

import (
	"context"

	"github.com/gdamore/tcell/v2"
	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
	"github.com/rivo/tview"
)

type AppState struct {
	userID         int64
	username       string
	currentTopicID int64
}

func Start(ctx context.Context, client pb.MessageBoardClient) {
	app := tview.NewApplication()
	state := &AppState{}

	showLoginScreen(ctx, app, client, state)

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func showLoginScreen(ctx context.Context, app *tview.Application, client pb.MessageBoardClient, state *AppState) {
	form := tview.NewForm()
	form.AddInputField("Username", "", 30, nil, nil)
	form.AddButton("Enter", func() {
		username := form.GetFormItemByLabel("Username").(*tview.InputField).GetText()
		if username == "" {
			return
		}

		resp, err := client.CreateUser(ctx, &pb.CreateUserRequest{Name: username})
		if err != nil {
			panic(err)
		}

		state.userID = resp.Id
		state.username = username
		showTopicsScreen(ctx, app, client, state)
	}).
		AddButton("Quit", func() {
			app.Stop()
		})

	form.SetBorder(true).SetTitle("Welcome to Message Board").SetTitleAlign(tview.AlignCenter)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			app.Stop()
		}
		return event
	})

	app.SetRoot(form, true).SetFocus(form)
}

func showTopicsScreen(ctx context.Context, app *tview.Application, client pb.MessageBoardClient, state *AppState) {
	list := tview.NewList().ShowSecondaryText(false)
	list.SetTitle("Topics").SetBorder(true).SetTitleAlign(tview.AlignCenter)

	// Helper to refresh the list
	updateList := func() {
		topicIDs := loadTopics(ctx, app, client, list, state)
		list.AddItem("", "", 0, nil)
		list.AddItem("+ Create new topic", "Press Enter to create", 'n', func() {
			showCreateTopicDialog(ctx, app, client, state)
		})
		// You might want to store topicIDs in a way that the SelectedFunc can access the latest version
		list.SetSelectedFunc(func(index int, _ string, _ string, _ rune) {
			if index >= len(topicIDs) {
				return
			}
			state.currentTopicID = topicIDs[index]
			showTopicScreen(ctx, app, client, state)
		})
	}

	updateList()

	infoView := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	infoView.SetText("[yellow]Enter:[-] Join | [yellow]N:[-] New Topic | [yellow]R:[-] Refresh | [yellow]Esc:[-] Quit")

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(list, 0, 1, true).
		AddItem(infoView, 1, 0, false)

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			app.Stop()
			return nil
		}
		if event.Rune() == 'r' || event.Rune() == 'R' {
			updateList()
			return nil
		}
		return event
	})

	app.SetRoot(layout, true).SetFocus(list)
}
