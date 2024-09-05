package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"strings"
	"time"
)

const (
	ClientHelloEcho          ClientMessageType = 1
	ClientInstructReply      ClientMessageType = 2
	ClientChromeDpScreenShot ClientMessageType = 3
	
	ServiceHelloEcho ServiceMessageType = 1
	ServiceInstruct  ServiceMessageType = 2
)

type MessageUseCase struct {
	instructRepo InstructRepo
	instanceRepo InstanceRepo
	logger       *zap.Logger
}

type ClientMessageType int32
type ClientMessage struct {
	Type               ClientMessageType `json:"type,omitempty"`
	Message            string            `json:"message"`
	InstructMessage    InstructMessage   `json:"instructMessage,omitempty"`
	ChromeDpScreenShot []byte            `json:"chromeDpScreenShot,omitempty"`
}

type ServiceMessageType int32
type ServiceMessage struct {
	Type            ServiceMessageType `json:"type,omitempty"`
	Message         string             `json:"message"`
	InstructMessage InstructMessage    `json:"instructMessage,omitempty"`
}

func NewMessageUseCase(logger *zap.Logger, instructRepo InstructRepo, instanceRepo InstanceRepo) *MessageUseCase {
	return &MessageUseCase{
		instructRepo: instructRepo,
		instanceRepo: instanceRepo,
		logger:       logger,
	}
}

// 定时发送 Server Hello 到发送消息通道

func (messageUseCase *MessageUseCase) HeartBeat(ctx context.Context, sendMsgChannel chan string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			//
			serviceMessage := ServiceMessage{
				Type:    ServiceHelloEcho,
				Message: fmt.Sprintf("Server Hello, Time: %s", time.Now().String()),
			}
			b, err := json.Marshal(serviceMessage)
			if err != nil {
				messageUseCase.logger.Error("发送HeartBeat消息至消息通道", zap.Error(err), zap.String("message", "序列化消息失败"))
				continue
			}
			
			sendMsgChannel <- string(b)
		
		case <-ctx.Done():
			messageUseCase.logger.Info("HeartBeat模块接收到关闭消息")
			return
		}
	}
	
}

// 接收socket消息，并传入 receiveMsgChannel 通道中

func (messageUseCase *MessageUseCase) ReceiveMessage(ctx context.Context, conn *websocket.Conn, receiveMsgChannel chan string, done chan struct{}) {
	for {
		// ReadMessage
		_, message, err := conn.ReadMessage()
		
		// 接收消息失败
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				messageUseCase.logger.Error("websocket已经关闭", zap.Error(err))
				close(done)
				return
			} else {
				messageUseCase.logger.Error("websocket错误", zap.Error(err))
				close(done)
				return
			}
		}
		
		//
		clientMsg := &ClientMessage{}
		
		err = json.Unmarshal(message, clientMsg)
		if err != nil {
			messageUseCase.logger.Error("反序列化客户端消息失败", zap.Error(err), zap.String("message", string(message)))
		} else {
			receiveMsgChannel <- string(message)
		}
	}
}

// 从 receiveMsgChannel 中获取消息并解析，然后执行对应的行为

func (messageUseCase *MessageUseCase) ProcessClientMessage(ctx context.Context, receiveMsgChannel chan string) {
	
	for {
		select {
		case m := <-receiveMsgChannel:
			
			var clientMsg ClientMessage
			err := json.Unmarshal([]byte(m), &clientMsg)
			if err != nil {
				messageUseCase.logger.Error("反序列化客户端消息失败", zap.Error(err))
				continue
			}
			
			//
			switch clientMsg.Type {
			case ClientHelloEcho:
				messageUseCase.logger.Info("接收到Client Hello消息",
					zap.String("message", clientMsg.Message),
				)
			
			case ClientInstructReply:
				switch clientMsg.InstructMessage.Type {
				case CommandInstruct:
					var result int32
					var reply string
					if clientMsg.InstructMessage.Result {
						result = 1
						reply = clientMsg.InstructMessage.CommandReply
					} else {
						result = -1
						reply = clientMsg.InstructMessage.ErrMsg
					}
					
					err = messageUseCase.instructRepo.UpdateInstruct(ctx, clientMsg.InstructMessage.Uuid, reply, result)
					if err != nil {
						messageUseCase.logger.Error("更新指令结果失败",
							zap.String("uuid", clientMsg.InstructMessage.Uuid),
							zap.Error(err))
					}
				
				case ChromeDpInspectInstruct:
					var result int32
					var reply string
					if clientMsg.InstructMessage.Result {
						result = 1
						reply = clientMsg.InstructMessage.ChromeDpInspectReply.String()
					} else {
						result = -1
						reply = clientMsg.InstructMessage.ErrMsg
					}
					
					err = messageUseCase.instructRepo.UpdateInstruct(ctx, clientMsg.InstructMessage.Uuid, reply, result)
					if err != nil {
						messageUseCase.logger.Error("更新指令结果失败",
							zap.String("uuid", clientMsg.InstructMessage.Uuid),
							zap.Error(err))
					}
				
				case DnsInstruct:
					var result int32
					var reply string
					if clientMsg.InstructMessage.Result {
						result = 1
						reply = strings.Join(clientMsg.InstructMessage.DnsInspectReply, ",")
					} else {
						result = -1
						reply = clientMsg.InstructMessage.ErrMsg
					}
					
					err = messageUseCase.instructRepo.UpdateInstruct(ctx, clientMsg.InstructMessage.Uuid, reply, result)
					if err != nil {
						messageUseCase.logger.Error("更新指令结果失败",
							zap.String("uuid", clientMsg.InstructMessage.Uuid),
							zap.Error(err))
					}
				
				case HttpInstruct:
					var result int32
					var reply string
					if clientMsg.InstructMessage.Result {
						result = 1
						reply = clientMsg.InstructMessage.HttpInspectReply.String()
					} else {
						result = -1
						reply = clientMsg.InstructMessage.ErrMsg
					}
					
					err = messageUseCase.instructRepo.UpdateInstruct(ctx, clientMsg.InstructMessage.Uuid, reply, result)
					if err != nil {
						messageUseCase.logger.Error("更新指令结果失败",
							zap.String("uuid", clientMsg.InstructMessage.Uuid),
							zap.Error(err))
					}
				
				case IcmpInstruct:
					var result int32
					var reply string
					if clientMsg.InstructMessage.Result {
						replyByte, err := json.Marshal(clientMsg.InstructMessage.IcmpInspectReply)
						if err != nil {
							reply = err.Error()
						} else {
							reply = string(replyByte)
						}
						result = 1
						//reply = clientMsg.InstructMessage.IcmpInspectReply
					} else {
						result = -1
						reply = clientMsg.InstructMessage.ErrMsg
					}
					
					err = messageUseCase.instructRepo.UpdateInstruct(ctx, clientMsg.InstructMessage.Uuid, reply, result)
					if err != nil {
						messageUseCase.logger.Error("更新指令结果失败",
							zap.String("uuid", clientMsg.InstructMessage.Uuid),
							zap.Error(err))
					}
				
				default:
					messageUseCase.logger.Error("未知的指令类型")
				}
				
				//messageUseCase.logger.Info("接收到Client指令响应消息",
				//	zap.String("message", clientMsg.InstructMessage.Reply),
				//)
			
			case ClientChromeDpScreenShot:
				messageUseCase.logger.Info("接收到Client ChromeDp截图消息",
					zap.String("message", string(clientMsg.ChromeDpScreenShot)),
				)
			
			default:
				messageUseCase.logger.Info("接收到未知客户端消息类型",
					zap.Any("Type", clientMsg.Type),
				)
			}
		
		case <-ctx.Done():
			return
		}
	}
}
