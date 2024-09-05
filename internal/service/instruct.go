package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/qx66/camp/internal/biz"
)

type InstructReq struct {
	OrgUuid      string `json:"orgUuid,omitempty" form:"orgUuid" validate:"required"`
	GroupUuid    string `json:"groupUuid,omitempty" form:"groupUuid" validate:"required"`
	InstanceName string `json:"instanceName,omitempty" form:"instanceName" validate:"required"`
	Type         int32  `json:"type,omitempty" form:"type" validate:"required"`
	Content      string `json:"content,omitempty" form:"content" validate:"required"`
}

func (useCase *UseCase) Instruct(c *gin.Context) {
	req := &InstructReq{}
	err := c.BindJSON(req)
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
	alive := useCase.instanceUseCase.GetInstanceAlive(c.Request.Context(), req.OrgUuid, req.GroupUuid, req.InstanceName)
	if !alive {
		c.JSON(400, gin.H{"errCode": 400, "errMsg": "目标实例不在线"})
		return
	}
	
	//
	instructUuid, err := useCase.instructUseCase.IssueInstructions(
		c.Request.Context(),
		req.OrgUuid,
		req.GroupUuid,
		req.InstanceName,
		biz.InstructType(req.Type),
		req.Content)
	
	if err != nil {
		c.JSON(500, gin.H{"errCode": 500, "errMsg": "系统异常"})
		return
	}
	
	c.JSON(200, gin.H{"errCode": 0, "errMsg": "ok", "uuid": instructUuid})
	return
}

// 列出指令

type ListInstructReq struct {
	OrgUuid      string `json:"orgUuid,omitempty" form:"orgUuid" validate:"required"`
	GroupUuid    string `json:"groupUuid,omitempty" form:"groupUuid" validate:"required"`
	InstanceName string `json:"instanceName,omitempty" form:"instanceName" validate:"required"`
	//Type         int32  `json:"type,omitempty" form:"type"`
}

func (useCase *UseCase) ListInstruct(c *gin.Context) {
	req := ListInstructReq{}
	err := c.ShouldBind(&req)
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
	instructs, err := useCase.instructUseCase.ListInstruct(c.Request.Context(), req.OrgUuid, req.GroupUuid, req.InstanceName)
	if err != nil {
		c.JSON(500, gin.H{"errCode": 500, "errMsg": "列出指令失败"})
		return
	}
	
	c.JSON(200, gin.H{"errCode": 0, "errMsg": "ok", "data": instructs})
}

// 获取指令信息

type GetInstructReq struct {
	OrgUuid      string `json:"orgUuid,omitempty" form:"orgUuid" validate:"required"`
	GroupUuid    string `json:"groupUuid,omitempty" form:"groupUuid" validate:"required"`
	InstanceName string `json:"instanceName,omitempty" form:"instanceName" validate:"required"`
}

func (useCase *UseCase) GetInstruct(c *gin.Context) {
	req := &GetInstructReq{}
	err := c.ShouldBind(&req)
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
	
	instructUuid := c.Param("uuid")
	
	//
	instruct, err := useCase.instructUseCase.GetInstruct(c.Request.Context(), req.OrgUuid, req.GroupUuid, req.InstanceName, instructUuid)
	if err != nil {
		c.JSON(500, gin.H{"errCode": 500, "errMsg": "列出指令失败"})
		return
	}
	
	c.JSON(200, gin.H{"errCode": 0, "errMsg": "ok", "data": instruct})
}
