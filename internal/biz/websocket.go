package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/http"
	"os/exec"
	"time"
)

type WebSocketUseCase struct {
	logger                   *zap.Logger
	dnsClientInspectUseCase  *DnsClientInspectUseCase
	httpInspectClientUseCase *HttpInspectClientUseCase
	icmpClientUseCase        *IcmpClientUseCase
	socketClientUseCase      *SocketClientUseCase
	chromeDpClientUseCase    *ChromeDpClientUseCase
}

func NewWebSocketUseCase(logger *zap.Logger,
	dnsClientInspectUseCase *DnsClientInspectUseCase,
	httpInspectClientUseCase *HttpInspectClientUseCase,
	icmpClientUseCase *IcmpClientUseCase,
	socketClientUseCase *SocketClientUseCase,
	chromeDpClientUseCase *ChromeDpClientUseCase) *WebSocketUseCase {
	return &WebSocketUseCase{
		logger:                   logger,
		dnsClientInspectUseCase:  dnsClientInspectUseCase,
		httpInspectClientUseCase: httpInspectClientUseCase,
		icmpClientUseCase:        icmpClientUseCase,
		socketClientUseCase:      socketClientUseCase,
		chromeDpClientUseCase:    chromeDpClientUseCase,
	}
}

// 创建 webSocket 连接

func (webSocketUseCase *WebSocketUseCase) connectWebSocket(wsUrl string, header http.Header) *websocket.Conn {
	var wsConn *websocket.Conn
	var err error
	
	for {
		wsConn, _, err = websocket.DefaultDialer.Dial(wsUrl, header)
		if err != nil {
			webSocketUseCase.logger.Error("连接服务器失败", zap.Error(err))
			time.Sleep(2 * time.Second)
			continue
		}
		
		break
	}
	
	webSocketUseCase.logger.Info("连接成功")
	return wsConn
}

func (webSocketUseCase *WebSocketUseCase) NewWebSocket(ctx context.Context, wsUrl string, token string, sendMsg chan ClientMessage, receiveMsg chan []byte, done chan struct{}) {
	
	header := http.Header{}
	header.Set("token", token)
	
	//
	var wsConn *websocket.Conn
	wsConn = webSocketUseCase.connectWebSocket(wsUrl, header)
	defer wsConn.Close()
	
	wsConn.SetPingHandler(func(appData string) error {
		err := wsConn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(1*time.Second))
		if err != nil {
			webSocketUseCase.logger.Error("Error sending Pong", zap.Error(err))
		}
		return nil
	})
	
	// read
	go func() {
		for {
			
			_, message, err := wsConn.ReadMessage()
			
			// 接收消息失败
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					webSocketUseCase.logger.Error("websocket已经关闭", zap.Error(err))
					
					wsConn = webSocketUseCase.connectWebSocket(wsUrl, header)
					//close(done)
					//return
				} else {
					webSocketUseCase.logger.Error("websocket错误", zap.Error(err))
					
					wsConn = webSocketUseCase.connectWebSocket(wsUrl, header)
					//close(done)
					//return
				}
			}
			
			receiveMsg <- message
			//app.logger.Info("接收服务端消息", zap.String("message", string(message)))
		}
	}()
	
	// write
	go func() {
		for {
			select {
			case msg := <-sendMsg:
				b, err := json.Marshal(msg)
				if err != nil {
					webSocketUseCase.logger.Error("序列化发送消息失败", zap.Error(err))
				} else {
					err = wsConn.WriteMessage(websocket.BinaryMessage, b)
					if err != nil {
						webSocketUseCase.logger.Error("发送websocket消息失败", zap.Error(err))
					}
				}
			
			case <-ctx.Done():
				wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				webSocketUseCase.logger.Info("写通道接收到关闭信号，将关闭通道")
				close(done)
				return
			}
		}
	}()
	
	select {
	case <-ctx.Done():
		webSocketUseCase.logger.Info("接收到关闭信号，程序退出")
		close(done)
		return
	}
}

// 处理服务端消息

func (webSocketUseCase *WebSocketUseCase) ProcessServiceMessage(ctx context.Context, serviceMsgChannel chan []byte, sendMsgChannel chan ClientMessage) {
	for {
		select {
		case serviceMsg := <-serviceMsgChannel:
			
			webSocketUseCase.logger.Info("处理服务器消息", zap.String("serviceMsg", string(serviceMsg)))
			
			serviceMessage := &ServiceMessage{}
			err := json.Unmarshal(serviceMsg, serviceMessage)
			if err != nil {
				webSocketUseCase.logger.Error("反序列化服务器消息失败", zap.Error(err), zap.String("message", string(serviceMsg)))
				continue
			}
			
			switch serviceMessage.Type {
			case ServiceHelloEcho:
				webSocketUseCase.logger.Info("服务器Echo消息", zap.String("message", serviceMessage.Message))
			
			case ServiceInstruct:
				
				switch serviceMessage.InstructMessage.Type {
				case CommandInstruct:
					var result bool
					var errMsg string
					//var replyContent string
					outPut, err := exec.Command("bash", "-c", serviceMessage.InstructMessage.CommandContent).Output()
					
					if err != nil {
						webSocketUseCase.logger.Error("执行命令行失败",
							zap.String("content", serviceMessage.InstructMessage.CommandContent),
							zap.Error(err))
						
						result = false
						errMsg = err.Error()
						//replyContent = fmt.Sprintf("执行命令行失败, output: %s, err: %s", string(outPut), err)
					} else {
						
						result = true
						errMsg = ""
						//replyContent = string(outPut)
					}
					
					reply := ClientMessage{
						Type: ClientInstructReply,
						InstructMessage: InstructMessage{
							Uuid:           serviceMessage.InstructMessage.Uuid,
							Type:           serviceMessage.InstructMessage.Type,
							CommandContent: serviceMessage.InstructMessage.CommandContent,
							CommandReply:   string(outPut),
							Result:         result,
							ErrMsg:         errMsg,
						},
					}
					
					sendMsgChannel <- reply
				
				case ChromeDpInspectInstruct:
					resp, err := webSocketUseCase.chromeDpClientUseCase.InspectSinglePage(ctx, serviceMessage.InstructMessage.ChromeDpInspectUrl)
					
					var result bool
					var errMsg string
					
					if err != nil {
						result = false
						errMsg = err.Error()
						
						webSocketUseCase.logger.Error("执行ChromeDp指令失败",
							zap.Error(err),
							zap.String("ChromeDpInspectUrl", serviceMessage.InstructMessage.ChromeDpInspectUrl),
						)
					} else {
						errMsg = ""
						result = true
						
						webSocketUseCase.logger.Info("执行ChromeDp指令成功",
							zap.String("ChromeDpInspectUrl", serviceMessage.InstructMessage.ChromeDpInspectUrl),
						)
					}
					
					reply := ClientMessage{
						Type: ClientInstructReply,
						InstructMessage: InstructMessage{
							Uuid:                 serviceMessage.InstructMessage.Uuid,
							Type:                 serviceMessage.InstructMessage.Type,
							ChromeDpInspectUrl:   serviceMessage.InstructMessage.ChromeDpInspectUrl,
							ChromeDpInspectReply: resp,
							Result:               result,
							ErrMsg:               errMsg,
						},
					}
					
					sendMsgChannel <- reply
				
				case DnsInstruct:
					resp, err := webSocketUseCase.dnsClientInspectUseCase.LookupHost(ctx, serviceMessage.InstructMessage.DnsInspectDomain)
					
					var result bool
					var errMsg string
					
					if err != nil {
						result = false
						errMsg = err.Error()
						
						webSocketUseCase.logger.Error("执行DNS指令失败",
							zap.Error(err),
							zap.String("DnsInspectDomain", serviceMessage.InstructMessage.DnsInspectDomain),
						)
					} else {
						errMsg = ""
						result = true
						
						webSocketUseCase.logger.Info("执行DNS指令成功",
							zap.String("ChromeDpInspectUrl", serviceMessage.InstructMessage.ChromeDpInspectUrl),
						)
					}
					
					reply := ClientMessage{
						Type: ClientInstructReply,
						InstructMessage: InstructMessage{
							Uuid:             serviceMessage.InstructMessage.Uuid,
							Type:             serviceMessage.InstructMessage.Type,
							DnsInspectDomain: serviceMessage.InstructMessage.DnsInspectDomain,
							DnsInspectReply:  resp,
							Result:           result,
							ErrMsg:           errMsg,
						},
					}
					
					sendMsgChannel <- reply
				
				case HttpInstruct:
					resp, err := webSocketUseCase.httpInspectClientUseCase.GetHttpUrlResponse(serviceMessage.InstructMessage.HttpInspectUrl)
					
					var result bool
					var errMsg string
					
					if err != nil {
						result = false
						errMsg = err.Error()
						
						webSocketUseCase.logger.Error("执行HTTP指令失败",
							zap.Error(err),
							zap.String("HttpInspectUrl", serviceMessage.InstructMessage.HttpInspectUrl),
						)
					} else {
						errMsg = ""
						result = true
						
						webSocketUseCase.logger.Info("执行HTTP指令成功",
							zap.String("HttpInspectUrl", serviceMessage.InstructMessage.HttpInspectUrl),
						)
					}
					
					reply := ClientMessage{
						Type: ClientInstructReply,
						InstructMessage: InstructMessage{
							Uuid:             serviceMessage.InstructMessage.Uuid,
							Type:             serviceMessage.InstructMessage.Type,
							HttpInspectUrl:   serviceMessage.InstructMessage.HttpInspectUrl,
							HttpInspectReply: resp,
							Result:           result,
							ErrMsg:           errMsg,
						},
					}
					
					sendMsgChannel <- reply
				
				case IcmpInstruct:
					resp, err := webSocketUseCase.icmpClientUseCase.Icmp(serviceMessage.InstructMessage.IcmpInspectAddr, 4)
					
					var result bool
					var errMsg string
					
					if err != nil {
						result = false
						errMsg = err.Error()
						
						webSocketUseCase.logger.Error("执行ICMP指令失败",
							zap.Error(err),
							zap.String("IcmpInspectAddr", serviceMessage.InstructMessage.IcmpInspectAddr),
						)
					} else {
						errMsg = ""
						result = true
						
						webSocketUseCase.logger.Info("执行ICMP指令成功",
							zap.String("IcmpInspectAddr", serviceMessage.InstructMessage.IcmpInspectAddr),
						)
					}
					
					reply := ClientMessage{
						Type: ClientInstructReply,
						InstructMessage: InstructMessage{
							Uuid:             serviceMessage.InstructMessage.Uuid,
							Type:             serviceMessage.InstructMessage.Type,
							IcmpInspectAddr:  serviceMessage.InstructMessage.IcmpInspectAddr,
							IcmpInspectReply: resp,
							Result:           result,
							ErrMsg:           errMsg,
						},
					}
					
					sendMsgChannel <- reply
				
				default:
					webSocketUseCase.logger.Warn("未知的指令",
						//zap.String("Content", serviceMessage.InstructMessage),
						zap.String("Uuid", serviceMessage.InstructMessage.Uuid),
						zap.Any("Type", serviceMessage.InstructMessage.Type),
					)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (webSocketUseCase *WebSocketUseCase) HelloEcho(ctx context.Context, sendMsg chan ClientMessage) {
	ticker := time.NewTicker(5 * time.Second)
	
	for {
		select {
		case <-ticker.C:
			//app.logger.Info("定时发送消息")
			sendMsg <- ClientMessage{
				Type:    ClientHelloEcho,
				Message: fmt.Sprintf("Hello, Client Time: %s", time.Now().String()),
			}
		
		case <-ctx.Done():
			webSocketUseCase.logger.Info("关闭定时发送消息")
			return
		}
	}
}
