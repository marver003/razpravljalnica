package ui

import (
	"context"
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
	"github.com/rivo/tview"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MessageWithUser struct {
	Message *pb.Message
	User    string
}

func loadTopics(ctx context.Context, app *tview.Application, client pb.MessageBoardClient, list *tview.List, state *AppState) []int64 {
	resp, err := client.ListTopics(ctx, &emptypb.Empty{})
	if err != nil {
		showErrorDialog(app, "Error while trying to load topics. Try again. ", func() {
			showTopicsScreen(ctx, app, client, state)
		})
		return nil
	}

	list.Clear()

	ids := make([]int64, 0, len(resp.Topics))
	for _, t := range resp.Topics {
		list.AddItem(fmt.Sprintf("%s", t.Name), "", 0, nil)
		ids = append(ids, t.Id)
	}

	return ids
}

func showErrorDialog(app *tview.Application, message string, onDismiss func()) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"V redu"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			onDismiss()
		})
	app.SetRoot(modal, true).SetFocus(modal)
}

func showCreateTopicDialog(ctx context.Context, app *tview.Application, client pb.MessageBoardClient, state *AppState) {
	form := tview.NewForm()
	form.AddInputField("Topic name", "", 30, nil, nil).
		AddButton("Create", func() {
			name := form.GetFormItemByLabel("Topic name").(*tview.InputField).GetText()
			if name == "" {
				return
			}

			_, err := client.CreateTopic(ctx, &pb.CreateTopicRequest{Name: name})
			if err != nil {
				showErrorDialog(app, "Napaka pri ustvarjanju topica: "+err.Error(), func() {
					showTopicsScreen(ctx, app, client, state)
				})
				return
			}

			showTopicsScreen(ctx, app, client, state)
		}).
		AddButton("Cancel", func() {
			showTopicsScreen(ctx, app, client, state)
		})

	form.SetBorder(true).SetTitle("Create New Topic").SetTitleAlign(tview.AlignCenter)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			showTopicsScreen(ctx, app, client, state)
		}
		return event
	})

	//log.Print("Here 000")

	app.SetRoot(form, true).SetFocus(form)
}

func showTopicScreen(ctx context.Context, app *tview.Application, client pb.MessageBoardClient, state *AppState) {
	// get topic name
	resp, err := client.ListTopics(ctx, &emptypb.Empty{})
	if err != nil {
		panic(err)
	}

	topicName := ""
	for _, t := range resp.Topics {
		if t.Id == state.currentTopicID {
			topicName = t.Name
			break
		}
	}

	// create message view
	messageView := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	messageView.SetTitle(fmt.Sprintf("Topic: %s", topicName)).SetBorder(true).SetTitleAlign(tview.AlignCenter)

	// create input field for sending messages
	inputField := tview.NewInputField().
		SetLabel("Message: ").
		SetFieldBackgroundColor(tcell.ColorDarkSlateGray).
		SetFieldTextColor(tcell.ColorWhite)

	// create info footer
	infoView := tview.NewTextView().SetDynamicColors(true).SetWrap(true).SetTextAlign(tview.AlignCenter)
	infoView.SetText("[yellow]Enter:[-] Send | [yellow]F2:[-] Like | [yellow]Esc:[-] Back | [yellow]↑↓:[-] Navigate")

	// store messages and selected index
	messages := make([]*pb.Message, 0)
	var messagesMutex sync.Mutex
	selectedMessageIndex := 0

	// refresh message display
	refreshMessages := func() {

		messageView.Clear()
		messagesMutex.Lock()
		defer messagesMutex.Unlock()

		for i, msg := range messages {
			marker := " "
			if i == selectedMessageIndex {
				marker = ">"
			}

			likeText := ""
			if msg.Likes > 0 {
				likeText = fmt.Sprintf(" [cyan]👍 %d[-]", msg.Likes)
			}

			fmt.Fprintf(messageView, "[%s] [yellow]User %d[-]: [white]%s[-]%s\n", marker, msg.UserId, msg.Text, likeText)
		}

		messageView.ScrollToEnd()

	}

	// load initial messages
	msgResp, err := client.GetMessages(ctx, &pb.GetMessagesRequest{
		TopicId:       state.currentTopicID,
		FromMessageId: 0,
		Limit:         100,
	})
	if err != nil {
		panic(err)
	}

	messagesMutex.Lock()
	messages = msgResp.Messages
	messagesMutex.Unlock()
	refreshMessages()

	// subscribe to new messages
	node, err := client.GetSubcscriptionNode(ctx, &pb.SubscriptionNodeRequest{
		UserId:  state.userID,
		TopicId: []int64{state.currentTopicID},
	})
	if err != nil {
		panic(err)
	}

	// connect to the specific node returned by the load balancer
	conn, err := grpc.NewClient(node.Node.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	subClient := pb.NewMessageBoardClient(conn)

	stream, err := subClient.SubscribeTopic(ctx, &pb.SubscribeTopicRequest{
		UserId:         state.userID,
		TopicId:        []int64{state.currentTopicID},
		FromMessageId:  0,
		SubscribeToken: node.SubscribeToken,
	})
	if err != nil {
		panic(err)
	}

	//log.Print("b")

	// listen for new events
	go func() {
		defer conn.Close()
		for {
			event, err := stream.Recv()
			if err != nil {
				return
			}

			// process data outside the UI thread
			messagesMutex.Lock()
			switch event.Op {
			case pb.OpType_OP_POST:
				messages = append(messages, event.Message)
			case pb.OpType_OP_LIKE:
				for i, msg := range messages {
					if msg.Id == event.Message.Id {
						messages[i] = event.Message
						break
					}
				}
			}
			messagesMutex.Unlock()

			app.QueueUpdateDraw(func() {
				refreshMessages()
			})
		}
	}()

	//log.Print("c")

	// handle input
	inputField.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}

		text := inputField.GetText()
		if text == "" {
			return
		}

		inputField.SetText("")

		go func() {
			_, err := client.PostMessage(ctx, &pb.PostMessageRequest{
				TopicId: state.currentTopicID,
				UserId:  state.userID,
				Text:    text,
			})

			if err != nil {
				panic(err)
			}
		}()
	})

	// setup layout
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(messageView, 0, 1, false).
		AddItem(inputField, 1, 0, true).
		AddItem(infoView, 1, 0, false)

	// setup input capture for message navigation and liking
	layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		messagesMutex.Lock()
		messageCount := len(messages)
		messagesMutex.Unlock()

		if messageCount == 0 {
			if event.Key() == tcell.KeyEsc {
				showTopicsScreen(ctx, app, client, state)
			}
			return event
		}

		switch {
		case event.Key() == tcell.KeyEsc:
			showTopicsScreen(ctx, app, client, state)
			return nil
		case event.Key() == tcell.KeyUp:
			if selectedMessageIndex > 0 {
				selectedMessageIndex--
				refreshMessages()
			}
			return nil
		case event.Key() == tcell.KeyDown:
			if selectedMessageIndex < messageCount-1 {
				selectedMessageIndex++
				refreshMessages()
			}
			return nil
		case event.Key() == tcell.KeyF2:
			messagesMutex.Lock()
			if selectedMessageIndex >= 0 && selectedMessageIndex < len(messages) {
				selectedMsg := messages[selectedMessageIndex]
				messagesMutex.Unlock()

				_, err := client.LikeMessage(ctx, &pb.LikeMessageRequest{
					TopicId:   state.currentTopicID,
					MessageId: selectedMsg.Id,
					UserId:    state.userID,
				})
				if err != nil {
					panic(err)
				}
			} else {
				messagesMutex.Unlock()
			}
			return nil
		}



		return event
	})

	app.SetRoot(layout, true).SetFocus(inputField)
}
