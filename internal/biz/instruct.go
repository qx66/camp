package biz

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/prometheus-community/pro-bing"
	"go.uber.org/zap"
	"time"
)

type InstructType int32

const (
	CommandInstruct         InstructType = 1 // 命令行指令
	ChromeDpInspectInstruct InstructType = 2 // ChromeDp任务指令
	DnsInstruct             InstructType = 3 // Dns 指令
	HttpInstruct            InstructType = 4 // Http 指令
	IcmpInstruct            InstructType = 5 // Mtr 指令
)

// DnsInspect: 为了方便管理，以及日常的需求，仅对域名进行 A记录 CName记录解析

type InstructMessage struct {
	Uuid           string       `json:"uuid,omitempty"`
	Type           InstructType `json:"type,omitempty"`
	CommandContent string       `json:"commandContent,omitempty"` // 命令行指令-命令
	CommandReply   string       `json:"commandReply,omitempty"`   // 命令行指令-返回内容
	
	ChromeDpInspectUrl   string                `json:"chromeDpInspectUrl,omitempty"`   // ChromeDp检测指令-Url
	ChromeDpInspectReply InspectSinglePageResp `json:"chromeDpInspectReply,omitempty"` // ChromeDp检测指令-返回内容
	
	DnsInspectDomain string   `json:"dnsContent,omitempty"` // Dns指令
	DnsInspectReply  []string `json:"dnsInspectReply,omitempty"`
	
	HttpInspectUrl   string           `json:"httpInspectUrl,omitempty"`   // Http指令
	HttpInspectReply HttpInspectReply `json:"httpInspectReply,omitempty"` // Http访问检测指令-返回内容
	
	IcmpInspectAddr  string `json:"icmpInspectAddr,omitempty"`
	IcmpInspectReply *probing.Statistics
	
	MtrInspectAddr  string `json:"mtrInspectAddr,omitempty"`  // mtr指令
	MtrInspectReply string `json:"mtrInspectReply,omitempty"` // mtr指令-返回内容
	Result          bool   `json:"result,omitempty"`          // 指令执行结果
	ErrMsg          string `json:"errMsg,omitempty"`
}

type Instruct struct {
	Uuid         string       `json:"uuid,omitempty"`
	OrgUuid      string       `json:"orgUuid,omitempty"`
	GroupUuid    string       `json:"groupUuid,omitempty"`
	InstanceName string       `json:"instanceName,omitempty"`
	Type         InstructType `json:"type,omitempty"`
	Content      string       `json:"content,omitempty"`
	Result       int32        `json:"result,omitempty"`
	Reply        string       `json:"reply,omitempty"`
	CreateTime   int64        `json:"createTime,omitempty"`
	UpdateTime   int64        `json:"updateTime,omitempty"`
}

func (instruct *Instruct) TableName() string {
	return "instruct"
}

type InstructRepo interface {
	IssueInstructions(ctx context.Context, orgUuid, groupUuid, instanceName string, instruct []byte) error
	ReceiveInstructions(ctx context.Context, orgUuid, groupUuid, instanceName string) (string, error)
	RecordInstruct(ctx context.Context, instruct Instruct) error
	UpdateInstruct(ctx context.Context, uuid, reply string, result int32) error
	ListInstruct(ctx context.Context, orgUuid, groupUuid, instanceName string) ([]Instruct, error)
	GetInstruct(ctx context.Context, orgUuid, groupUuid, instanceName, uuid string) (Instruct, error)
}

type InstructUseCase struct {
	instructRepo InstructRepo
	logger       *zap.Logger
}

func NewInstructUseCase(instructRepo InstructRepo, logger *zap.Logger) *InstructUseCase {
	return &InstructUseCase{
		instructRepo: instructRepo,
		logger:       logger,
	}
}

// 发布指令

func (instructUseCase *InstructUseCase) IssueInstructions(ctx context.Context, orgUuid string, groupUuid string, instanceName string, instructType InstructType, instructContent string) (string, error) {
	instructUuid := uuid.NewString()
	var serviceMessage ServiceMessage
	
	switch instructType {
	case CommandInstruct:
		serviceMessage = ServiceMessage{
			Type: ServiceInstruct,
			InstructMessage: InstructMessage{
				Uuid:           instructUuid,
				Type:           instructType,
				CommandContent: instructContent,
			},
		}
	case ChromeDpInspectInstruct:
		serviceMessage = ServiceMessage{
			Type: ServiceInstruct,
			InstructMessage: InstructMessage{
				Uuid:               instructUuid,
				Type:               instructType,
				ChromeDpInspectUrl: instructContent,
			},
		}
		//return errors.New("暂不支持")
	
	case DnsInstruct:
		serviceMessage = ServiceMessage{
			Type: ServiceInstruct,
			InstructMessage: InstructMessage{
				Uuid:             instructUuid,
				Type:             instructType,
				DnsInspectDomain: instructContent,
			},
		}
		//return instructUuid, errors.New("暂不支持")
	
	case HttpInstruct:
		serviceMessage = ServiceMessage{
			Type: ServiceInstruct,
			InstructMessage: InstructMessage{
				Uuid:           instructUuid,
				Type:           instructType,
				HttpInspectUrl: instructContent,
			},
		}
		//return instructUuid, errors.New("暂不支持")
	
	case IcmpInstruct:
		serviceMessage = ServiceMessage{
			Type: ServiceInstruct,
			InstructMessage: InstructMessage{
				Uuid:            instructUuid,
				Type:            instructType,
				IcmpInspectAddr: instructContent,
			},
		}
		//return instructUuid, errors.New("暂不支持")
	
	default:
		instructUseCase.logger.Error("未知指令类型", zap.Any("instructType", instructType))
		return instructUuid, errors.New("未知指令类型")
	}
	
	jsonByte, err := json.Marshal(serviceMessage)
	if err != nil {
		return instructUuid, err
	}
	
	err = instructUseCase.instructRepo.RecordInstruct(ctx, Instruct{
		Uuid:         instructUuid,
		OrgUuid:      orgUuid,
		GroupUuid:    groupUuid,
		InstanceName: instanceName,
		Type:         instructType,
		Content:      instructContent,
		Result:       0,
		Reply:        "",
		CreateTime:   time.Now().Unix(),
		UpdateTime:   time.Now().Unix(),
	})
	if err != nil {
		return instructUuid, errors.New("记录数据到数据库失败")
	}
	
	return instructUuid, instructUseCase.instructRepo.IssueInstructions(ctx, orgUuid, groupUuid, instanceName, jsonByte)
}

// 接收指令 - 采用 redis list 存储指令，建议每秒获取一次

func (instructUseCase *InstructUseCase) ReceiveInstructions(ctx context.Context, orgUuid string, groupUuid string, instanceName string, instructions chan string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			//instructUseCase.logger.Info("定时执行接收到指令消息")
			instruction, err := instructUseCase.instructRepo.ReceiveInstructions(ctx, orgUuid, groupUuid, instanceName)
			if err == nil {
				//instructUseCase.logger.Info("接收到指令消息", zap.String("instruction", instruction))
				instructions <- instruction
			}
		case <-ctx.Done():
			instructUseCase.logger.Info("接收到指令消息结束")
			return
		}
	}
}

// 列出指令

func (instructUseCase *InstructUseCase) ListInstruct(ctx context.Context, orgUuid, groupUuid, instanceName string) ([]Instruct, error) {
	return instructUseCase.instructRepo.ListInstruct(ctx, orgUuid, groupUuid, instanceName)
}

// 获取指令

func (instructUseCase *InstructUseCase) GetInstruct(ctx context.Context, orgUuid, groupUuid, instanceName, uuid string) (Instruct, error) {
	return instructUseCase.instructRepo.GetInstruct(ctx, orgUuid, groupUuid, instanceName, uuid)
}
