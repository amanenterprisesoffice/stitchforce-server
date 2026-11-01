
package request

type CreateSize struct {
    Name string `form:"name" json:"name" binding:"required,min=1,max=50"`
}

type UpdateSize struct {
    Name string `form:"name" json:"name" binding:"required,min=1,max=50"`
}
