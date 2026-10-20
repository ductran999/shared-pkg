package response

// Pagination holds pagination metadata for list responses.
type Pagination struct {
	Page  int `json:"page"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// NewPagination builds pagination metadata from page, size and total.
//
// Example:
//
//	return response.List(items, response.NewPagination(page, size, total))
func NewPagination(page, size, total int) Pagination {
	return Pagination{Page: page, Size: size, Total: total}
}
