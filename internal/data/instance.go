package data

import (
	"context"
	"fmt"
	"github.com/qx66/camp/internal/biz"
	"gorm.io/gorm"
	"time"
)

type InstanceDataSource struct {
	data *Data
}

func NewInstanceDataSource(data *Data) biz.InstanceRepo {
	return &InstanceDataSource{
		data: data,
	}
}

func (instanceDataSource *InstanceDataSource) HeartBeat(ctx context.Context, orgUuid string, groupUuid string, instanceName string) error {
	key := fmt.Sprintf("%s_%s_%s_heartbeat", orgUuid, groupUuid, instanceName)
	return instanceDataSource.data.redis.Set(ctx, key, "", 10*time.Second).Err()
}

func (instanceDataSource *InstanceDataSource) Register(ctx context.Context, instance biz.Instance) (string, error) {
	//
	iInstance := biz.Instance{}
	
	tx := instanceDataSource.data.db.
		WithContext(ctx).
		Where("org_uuid = ? and group_uuid = ? and instance_name = ?",
			instance.OrgUuid, instance.GroupUuid, instance.InstanceName).
		First(&iInstance)
	
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			tx = instanceDataSource.data.db.WithContext(ctx).Create(&instance)
			return instance.Uuid, tx.Error
		}
		return iInstance.Uuid, tx.Error
	}
	
	tx = instanceDataSource.data.db.Model(&biz.Instance{}).
		WithContext(ctx).
		Where("org_uuid = ? and group_uuid = ? and instance_name = ?",
			instance.OrgUuid, instance.GroupUuid, instance.InstanceName).
		Updates(map[string]interface{}{
			"client_ip":   instance.ClientIp,
			"create_time": instance.CreateTime,
		})
	return iInstance.Uuid, tx.Error
}

func (instanceDataSource *InstanceDataSource) List(ctx context.Context, orgUuid string, groupUuid string) ([]biz.Instance, error) {
	var instances []biz.Instance
	tx := instanceDataSource.data.db.WithContext(ctx).Where("org_uuid = ?", orgUuid)
	if groupUuid != "" {
		tx.Where("group_uuid = ?", groupUuid)
	}
	
	tx.Order("create_time desc").Find(&instances)
	return instances, tx.Error
}

func (instanceDataSource *InstanceDataSource) ListAliveInstance(ctx context.Context, orgUuid string, groupUuid string) ([]biz.Instance, error) {
	var instances []biz.Instance
	
	tx := instanceDataSource.data.db.WithContext(ctx).Where("update_time >= ?", time.Now().Unix()-20)
	
	if orgUuid != "" {
		tx.Where("org_uuid = ?", orgUuid)
	}
	
	if groupUuid != "" {
		tx.Where("group_uuid = ?", groupUuid)
	}
	
	tx.Order("create_time desc").Find(&instances)
	return instances, tx.Error
}

func (instanceDataSource *InstanceDataSource) Get(ctx context.Context, orgUuid string, groupUuid string, instanceName string) (biz.Instance, error) {
	var instance biz.Instance
	
	tx := instanceDataSource.data.db.WithContext(ctx).
		Where("org_uuid = ? and group_uuid = ? and instance_name = ?", orgUuid, groupUuid, instanceName).
		First(&instance)
	return instance, tx.Error
}

func (instanceDataSource *InstanceDataSource) UpdateTime(ctx context.Context, uuid string) error {
	tx := instanceDataSource.data.db.WithContext(ctx).
		Model(&biz.Instance{}).
		Where("uuid = ?", uuid).
		Update("update_time", time.Now().Unix())
	return tx.Error
}
