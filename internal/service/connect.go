package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/websocket"
	"github.com/qx66/camp/pkg/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ConnectReq struct {
	OrgUuid      string `json:"orgUuid,omitempty" form:"orgUuid" validate:"required"`
	GroupUuid    string `json:"groupUuid,omitempty" form:"groupUuid" validate:"required"`
	InstanceName string `json:"instanceName,omitempty" form:"instanceName" validate:"required"`
}

func (useCase *UseCase) Connect(c *gin.Context) {
	//useCase.logger.Info("c.Request.Header", zap.Any("c.Request.Header", c.Request.Header))
	clientIp := middleware.GetClientIp(c)
	useCase.logger.Info("clientIp", zap.String("clientIp", clientIp))
	
	// 1. 参数处理
	req := &ConnectReq{}
	err := c.BindQuery(req)
	if err != nil {
		c.JSON(400, gin.H{"errCode": 400, "errMsg": "参数异常"})
		return
	}
	
	//
	validate := validator.New()
	err = validate.Struct(req)
	if err != nil {
		c.JSON(400, gin.H{"errCode": 400, "errMsg": "参数异常"})
		return
	}
	
	// 2. 协议升级
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.Set("error", err.Error())
		c.JSON(500, gin.H{"errCode": 500, "errMsg": "Internal Server Error"})
		return
	}
	defer conn.Close()
	
	conn.SetPingHandler(func(appData string) error {
		err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(1*time.Second))
		if err != nil {
			useCase.logger.Error("Error sending Pong", zap.Error(err))
			return err
		}
		useCase.logger.Info("发送心跳成功")
		return nil
	})
	
	//
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	
	useCase.logger.Info("用户连接成功",
		zap.String("orgUuid", req.OrgUuid),
		zap.String("groupUuid", req.GroupUuid),
		zap.String("instanceName", req.InstanceName),
	)
	
	// message channel
	receiveMsgChannel := make(chan string)
	sendMsgChannel := make(chan string, 10)
	done := make(chan struct{})
	
	//
	// 4.1 接收消息
	go useCase.messageUseCase.ReceiveMessage(ctx, conn, receiveMsgChannel, done)
	// 4.3 处理消息
	go useCase.messageUseCase.ProcessClientMessage(ctx, receiveMsgChannel)
	
	go useCase.instanceUseCase.UpdateTime(ctx, req.OrgUuid, req.GroupUuid, req.InstanceName, clientIp)
	
	// 4.2 定时心跳
	// go useCase.messageUseCase.HeartBeat(ctx, sendMsgChannel)
	
	// 4.4 将指令消息发送到消息通道
	go useCase.instructUseCase.ReceiveInstructions(ctx, req.OrgUuid, req.GroupUuid, req.InstanceName, sendMsgChannel)
	
	// 4.5 从 sendMsgChannel 中获取数据并发送
	go func() {
		for {
			select {
			case m, ok := <-sendMsgChannel:
				if !ok {
					useCase.logger.Info("获取发送通道异常，通道已关闭",
						zap.String("orgUuid", req.OrgUuid),
						zap.String("groupUuid", req.GroupUuid),
						zap.String("instanceName", req.InstanceName),
					)
					return
				}
				
				err = conn.WriteMessage(websocket.TextMessage, []byte(m))
				if err != nil {
					useCase.logger.Error("发送消息给客户端失败", zap.Error(err))
				} else {
					useCase.logger.Info("发送消息给客户端",
						zap.String("orgUuid", req.OrgUuid),
						zap.String("groupUuid", req.GroupUuid),
						zap.String("instanceName", req.InstanceName),
						zap.String("message", m))
				}
			
			case <-ctx.Done():
				useCase.logger.Info("websocket is closed")
				cancel()
				return
			}
		}
	}()
	
	// 5. done
	select {
	case <-ctx.Done():
		useCase.logger.Info("websocket is closed")
		cancel()
		return
	case <-done:
		useCase.logger.Info("websocket is closed")
		cancel()
		return
	}
	
}
