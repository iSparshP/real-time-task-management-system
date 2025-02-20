package task

import "time"

type TaskFilter struct {
	Status     *string    `form:"status"`
	Priority   *string    `form:"priority"`
	AssignedTo *string    `form:"assigned_to"`
	CreatedBy  *string    `form:"created_by"`
	DueBefore  *time.Time `form:"due_before"`
	DueAfter   *time.Time `form:"due_after"`
}

type PaginationParams struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=10"`
}

type SortParams struct {
	SortBy    string `form:"sort_by,default=created_at"`
	SortOrder string `form:"sort_order,default=desc"`
}
