package request

import "stitchforce-server/internal/entity"

type CreateGarmentWorker struct {
	Name   string `form:"name" json:"name" binding:"required,min=2,max=200"`
	Phone  string `form:"phone" json:"phone" binding:"required,min=10,max=15"`
	Gender string `form:"gender" json:"gender" binding:"required,oneof=male female other"`
}

type UpdateGarmentWorker struct {
	Name   *string              `form:"name" json:"name" binding:"omitempty,min=2,max=200"`
	Phone  *string              `form:"phone" json:"phone" binding:"omitempty,min=10,max=15"`
	Gender *string              `form:"gender" json:"gender" binding:"omitempty,oneof=male female other"`
	Status *entity.WorkerStatus `form:"status" json:"status" binding:"omitempty,oneof=active inactive"`
}