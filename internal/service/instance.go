package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ListAliveInstanceReq struct {
	OrgUuid   string `json:"orgUuid,omitempty" form:"orgUuid"`
	GroupUuid string `json:"groupUuid,omitempty" form:"groupUuid"`
}

func (useCase *UseCase) ListAliveInstance(c *gin.Context) {
	// 1. 参数处理
	req := &ListAliveInstanceReq{}
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
	
	//
	instances, err := useCase.instanceUseCase.ListAliveInstance(c.Request.Context(), req.OrgUuid, req.GroupUuid)
	if err != nil {
		c.JSON(500, gin.H{"errCode": 500, "errMsg": "系统异常"})
		return
	}
	
	c.JSON(200, gin.H{"errCode": 0, "errMsg": "ok", "instances": instances})
	return
}
