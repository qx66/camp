package biz

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

type Instance struct {
	Uuid         string `json:"uuid,omitempty"`
	OrgUuid      string `json:"orgUuid,omitempty"`
	GroupUuid    string `json:"groupUuid,omitempty"`
	InstanceName string `json:"instanceName,omitempty"`
	ClientIp     string `json:"clientIp,omitempty"`
	CreateTime   int64  `json:"createTime,omitempty"`
	UpdateTime   int64  `json:"updateTime,omitempty"`
}

func (instance *Instance) TableName() string {
	return "instance"
}

type InstanceRepo interface {
	Register(ctx context.Context, instance Instance) (string, error)
	List(ctx context.Context, orgUuid, groupUuid string) ([]Instance, error)
	ListAliveInstance(ctx context.Context, orgUuid, groupUuid string) ([]Instance, error)
	Get(ctx context.Context, orgUuid, groupUuid, instanceName string) (Instance, error)
	HeartBeat(ctx context.Context, orgUuid string, groupUuid string, instanceName string) error
	UpdateTime(ctx context.Context, uuid string) error
}

type InstanceUseCase struct {
	instanceRepo InstanceRepo
	logger       *zap.Logger
}

func NewInstanceUseCase(instanceRepo InstanceRepo, logger *zap.Logger) *InstanceUseCase {
	return &InstanceUseCase{
		instanceRepo: instanceRepo,
		logger:       logger,
	}
}

func (instanceUseCase *InstanceUseCase) Register(ctx context.Context, orgUuid, groupUuid, instanceName, clientIp string) (string, error) {
	uid := uuid.NewString()
	instance := Instance{
		Uuid:         uid,
		OrgUuid:      orgUuid,
		GroupUuid:    groupUuid,
		InstanceName: instanceName,
		ClientIp:     clientIp,
		CreateTime:   time.Now().Unix(),
		UpdateTime:   time.Now().Unix(),
	}
	
	return instanceUseCase.instanceRepo.Register(ctx, instance)
}

// 定时更新状态

func (instanceUseCase *InstanceUseCase) UpdateTime(ctx context.Context, orgUuid, groupUuid, instanceName, clientIp string) {
	//
	var uid string
	var err error
	
	for {
		uid, err = instanceUseCase.Register(ctx, orgUuid, groupUuid, instanceName, clientIp)
		if err == nil {
			break
		} else {
			instanceUseCase.logger.Error("注册实例失败", zap.Error(err))
		}
		time.Sleep(3 * time.Second)
	}
	
	//
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			//
			err = instanceUseCase.instanceRepo.UpdateTime(ctx, uid)
			if err != nil {
				instanceUseCase.logger.Error("更新实例时间失败", zap.String("uuid", uid), zap.Error(err))
			}
		
		case <-ctx.Done():
			instanceUseCase.logger.Info("定时更新时间模块关闭消息")
			return
		}
	}
	
}

//

func (instanceUseCase *InstanceUseCase) ListAliveInstance(ctx context.Context, orgUuid, groupUuid string) ([]Instance, error) {
	instanceUseCase.logger.Info("ListAliveInstance", zap.String("orgUuid", orgUuid), zap.String("groupUuid", groupUuid))
	return instanceUseCase.instanceRepo.ListAliveInstance(ctx, orgUuid, groupUuid)
}

//

func (instanceUseCase *InstanceUseCase) GetInstanceAlive(ctx context.Context, orgUuid, groupUuid, instanceName string) bool {
	instance, err := instanceUseCase.instanceRepo.Get(ctx, orgUuid, groupUuid, instanceName)
	if err != nil {
		return false
	}
	
	if (time.Now().Unix() - instance.UpdateTime) >= 20 {
		return false
	}
	
	return true
}
