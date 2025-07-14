package dto

// Response 通用响应结构
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Pagination 分页信息
type Pagination struct {
	Total    int64 `json:"total"`     // 总记录数
	Page     int   `json:"page"`      // 当前页码
	PageSize int   `json:"page_size"` // 每页数量
	Pages    int   `json:"pages"`     // 总页数
}

// NewPagination 创建分页信息
func NewPagination(total int64, page, pageSize int) Pagination {
	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}
	return Pagination{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	}
}
