package data

import (
	"context"
	"fmt"
	"github.com/qx66/camp/internal/biz"
	"time"
)

type InstructDataSource struct {
	data *Data
}

func NewInstructDataSource(data *Data) biz.InstructRepo {
	return &InstructDataSource{
		data: data,
	}
}

func (instructDataSource *InstructDataSource) IssueInstructions(ctx context.Context, orgUuid string, groupUuid string, instanceName string, instruct []byte) error {
	key := fmt.Sprintf("%s_%s_%s_instructions_channel", orgUuid, groupUuid, instanceName)
	return instructDataSource.data.redis.LPush(ctx, key, instruct).Err()
}

func (instructDataSource *InstructDataSource) ReceiveInstructions(ctx context.Context, orgUuid string, groupUuid string, instanceName string) (string, error) {
	key := fmt.Sprintf("%s_%s_%s_instructions_channel", orgUuid, groupUuid, instanceName)
	return instructDataSource.data.redis.LPop(ctx, key).Result()
}

func (instructDataSource *InstructDataSource) RecordInstruct(ctx context.Context, instruct biz.Instruct) error {
	tx := instructDataSource.data.db.WithContext(ctx).Create(&instruct)
	return tx.Error
}

func (instructDataSource *InstructDataSource) UpdateInstruct(ctx context.Context, uuid, reply string, result int32) error {
	tx := instructDataSource.data.db.WithContext(ctx).
		Model(&biz.Instruct{}).
		Where("uuid = ?", uuid).
		Updates(map[string]interface{}{
			"reply":       reply,
			"update_time": time.Now().Unix(),
			"result":      result,
		})
	
	return tx.Error
}

func (instructDataSource *InstructDataSource) ListInstruct(ctx context.Context, orgUuid string, groupUuid string, instanceName string) ([]biz.Instruct, error) {
	var instructs []biz.Instruct
	tx := instructDataSource.data.db.WithContext(ctx).
		Where("org_uuid = ? and group_uuid = ? and instance_name = ?", orgUuid, groupUuid, instanceName).
		Order("create_time desc").
		Limit(50).
		Find(&instructs)
	return instructs, tx.Error
}

func (instructDataSource *InstructDataSource) GetInstruct(ctx context.Context, orgUuid string, groupUuid string, instanceName string, uuid string) (biz.Instruct, error) {
	var instruct biz.Instruct
	tx := instructDataSource.data.db.WithContext(ctx).
		Where("org_uuid = ? and group_uuid = ? and instance_name = ? and uuid = ?", orgUuid, groupUuid, instanceName, uuid).
		First(&instruct)
	return instruct, tx.Error
}
