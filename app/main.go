package main

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
	"github.com/qx66/camp/internal/biz"
	"go.uber.org/zap"
	"net/url"
	
	//lTheme "github.com/qx66/camp/pkg/theme"
)

type iApp struct {
	logger           *zap.Logger
	webSocketUseCase *biz.WebSocketUseCase
}

func newIApp(logger *zap.Logger, webSocketUseCase *biz.WebSocketUseCase) *iApp {
	return &iApp{
		logger:           logger,
		webSocketUseCase: webSocketUseCase,
	}
}

var (
	webSocketUrl = "ws://camp.startops.com.cn/connect"
	token        = ""
	orgUuid      = "mobile"
	groupUuid    = "android"
	instanceName = ""
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println("新建logger失败")
		return
	}
	
	instanceName = uuid.NewString()
	logger.Info("实例信息", zap.String("orgUuid", orgUuid), zap.String("groupUuid", groupUuid), zap.String("instanceName", instanceName))
	
	ctx := context.Background()
	iApp := initApp(logger)
	
	websocketUrl := fmt.Sprintf("%s?orgUuid=%s&groupUuid=%s&instanceName=%s",
		webSocketUrl, orgUuid, groupUuid, instanceName)
	
	sendMsg := make(chan biz.ClientMessage)
	receiveMsg := make(chan []byte)
	done := make(chan struct{})
	defer close(done)
	
	go iApp.webSocketUseCase.NewWebSocket(ctx, websocketUrl, token, sendMsg, receiveMsg, done)
	
	go iApp.webSocketUseCase.ProcessServiceMessage(ctx, receiveMsg, sendMsg)
	
	go iApp.webSocketUseCase.HelloEcho(ctx, sendMsg)
	
	myApp := app.New()
	//myApp.Settings().SetTheme(&lTheme.MyTheme{})
	win := myApp.NewWindow("Camp")
	win.SetContent(iApp.makeUI(win))
	win.Resize(fyne.NewSize(win.Canvas().Size().Width, 650))
	win.ShowAndRun()
}

func (iApp *iApp) makeUI(w fyne.Window) fyne.CanvasObject {
	//ctx := context.Background()
	// header
	header := canvas.NewText("Camp", theme.PrimaryColor())
	header.TextSize = 42
	header.Alignment = fyne.TextAlignCenter
	
	// foot
	u, _ := url.Parse("https://www.startops.com.cn")
	footer := widget.NewHyperlinkWithStyle("www.startops.com.cn", u, fyne.TextAlignCenter, fyne.TextStyle{})
	
	//
	input := widget.NewEntry()
	input.MultiLine = true
	input.Wrapping = fyne.TextWrapBreak
	input.SetPlaceHolder(fmt.Sprintf("instanceName: %s", instanceName))
	
	output := widget.NewEntry()
	output.MultiLine = true
	output.Wrapping = fyne.TextWrapBreak
	output.SetPlaceHolder("Output Result")
	
	/*
		diag := widget.NewButtonWithIcon(
			"诊断",
			theme.MediaSkipNextIcon(),
			func() {
				if input.Text == "" {
					//input.Text = w.Clipboard().Content()
					input.Refresh()
				}
	
				output.Refresh()
			})
	
		diag.Importance = widget.HighImportance
	*/
	
	/*
		clear := widget.NewButtonWithIcon(
			"clear",
			theme.ContentClearIcon(),
			func() {
				output.Text = ""
				output.Refresh()
				input.Text = ""
				input.Refresh()
			},
		)
		clear.Importance = widget.MediumImportance
	*/
	
	copy := widget.NewButtonWithIcon(
		"拷贝实例名",
		theme.ContentCutIcon(),
		func() {
			clipboard := w.Clipboard()
			clipboard.SetContent(instanceName)
			//output.Text = ""
			//output.Refresh()
			//input.Text = ""
			//input.Refresh()
		},
	)
	copy.Importance = widget.WarningImportance
	
	return container.NewBorder(
		header,
		footer,
		nil,
		nil,
		container.NewGridWithRows(
			2,
			container.NewBorder(
				nil,
				//container.NewVBox(container.NewGridWithColumns(3, diag, clear, decode), copy),
				//container.NewVBox(container.NewGridWithColumns(3, diag, clear, copy)),
				container.NewVBox(container.NewGridWithColumns(1, copy)),
				nil,
				nil,
				input),
			output,
		),
	)
}
