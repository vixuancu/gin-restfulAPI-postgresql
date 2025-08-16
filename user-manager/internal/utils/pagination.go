package utils

import "strconv"

type Pagination struct {
	Page         int32 `json:"page"`
	Limit        int32 `json:"limit"`
	TotalRecords int32 `json:"total_records"`
	TotalPages   int32 `json:"total_pages"`
	HasNext      bool  `json:"has_next"`
	HasPrev      bool  `json:"has_prev"`
}

func NewPagination(page, limit, totalRecords int32) *Pagination {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 500 {
		envLimit := GetEnv("LIMIT_ITEM_ON_PER_PAGE", "10") // Lấy giá trị từ biến môi trường, nếu không có thì mặc định là 10
		limitInt, err := strconv.Atoi(envLimit)            // Chuyển đổi chuỗi sang số nguyên
		if err != nil && limitInt < 1 {
			limitInt = 10 // Nếu không thể chuyển đổi hoặc giá trị nhỏ hơn 1 thì mặc định là 10
		}
		limit = int32(limitInt) // Cập nhật giá trị limit
	}
	totalPages := (totalRecords + limit - 1) / limit // Tính tổng số trang, làm tròn lên
	return &Pagination{
		Page:         page,
		Limit:        limit,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,        // Tính tổng số trang
		HasNext:      page < totalPages, // Kiểm tra xem có trang tiếp theo không
		HasPrev:      page > 1,          // Kiểm tra xem có trang trước không
	}
}
func NewPaginationResponse(data any, page, limit, totalRecords int32) map[string]any {
	return map[string]any{
		"data":       data,
		"pagination": NewPagination(page, limit, totalRecords),
	}

}
