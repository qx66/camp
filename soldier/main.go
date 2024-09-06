package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/qx66/camp/internal/biz"
	"go.uber.org/zap"
	"os"
	"os/signal"
)

type app struct {
	logger *zap.Logger
	//chromeDpClientUseCase    *biz.ChromeDpClientUseCase
	//dnsClientInspectUseCase  *biz.DnsClientInspectUseCase
	//httpInspectClientUseCase *biz.HttpInspectClientUseCase
	//icmpClientUseCase        *biz.IcmpClientUseCase
	//socketClientUseCase      *biz.SocketClientUseCase
	webSocketUseCase *biz.WebSocketUseCase
}

func newApp(logger *zap.Logger,
//chromeDpClientUseCase *biz.ChromeDpClientUseCase,
//dnsClientInspectUseCase *biz.DnsClientInspectUseCase,
//httpInspectClientUseCase *biz.HttpInspectClientUseCase,
//icmpClientUseCase *biz.IcmpClientUseCase,
//socketClientUseCase *biz.SocketClientUseCase,
	webSocketUseCase *biz.WebSocketUseCase) *app {
	return &app{
		logger: logger,
		//chromeDpClientUseCase:    chromeDpClientUseCase,
		//dnsClientInspectUseCase:  dnsClientInspectUseCase,
		//httpInspectClientUseCase: httpInspectClientUseCase,
		//icmpClientUseCase:        icmpClientUseCase,
		//socketClientUseCase:      socketClientUseCase,
		webSocketUseCase: webSocketUseCase,
	}
}

var (
	webSocketUrl = ""
	token        = ""
	orgUuid      = ""
	groupUuid    = ""
	instanceName = ""
)

func init() {
	flag.StringVar(&webSocketUrl, "webSocketUrl", "", "websocket Url (required), e.g: ws://camp.startops.com.cn/connect")
	flag.StringVar(&token, "token", "", "websocket token (required)")
	flag.StringVar(&orgUuid, "orgUuid", "", "your orgUuid (required)")
	flag.StringVar(&groupUuid, "groupUuid", "", "your groupUuid (required)")
	flag.StringVar(&instanceName, "instanceName", "", "your instanceName (required)")
}

func main() {
	flag.Parse()
	
	logger, err := zap.NewProduction()
	if err != nil {
		panic("新建logger失败")
		return
	}
	
	if webSocketUrl == "" || orgUuid == "" || groupUuid == "" || instanceName == "" {
		logger.Error("请输入参数")
		return
	}
	
	websocketUrl := fmt.Sprintf("%s?orgUuid=%s&groupUuid=%s&instanceName=%s",
		webSocketUrl, orgUuid, groupUuid, instanceName)
	
	//
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	//
	sendMsg := make(chan biz.ClientMessage)
	receiveMsg := make(chan []byte)
	done := make(chan struct{})
	//actionChannel := make(chan Action)
	
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	
	app := initApp(logger)
	
	go app.webSocketUseCase.NewWebSocket(ctx, websocketUrl, token, sendMsg, receiveMsg, done)
	
	go app.webSocketUseCase.ProcessServiceMessage(ctx, receiveMsg, sendMsg)
	
	go app.webSocketUseCase.HelloEcho(ctx, sendMsg)
	
	select {
	case <-sig:
		cancel()
		return
	case <-done:
		cancel()
		return
	}
}
