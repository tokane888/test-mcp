package query

type ListUsers struct {
	Limit  int `form:"limit,default=10" binding:"min=1,max=100"`
	Offset int `form:"offset,default=0" binding:"min=0"`
}
