package handler

import (
	"errors"
	"strconv"

	"github.com/tunabsdrmz/boxing-gym-management/internal/repository"
)

// ResolveListPagination parses limit and either page (1-based) or offset.
// If pageStr is non-empty, offset is (page-1)*limit and offset query is ignored.
// If pageStr is empty, offsetStr is used (default 0).
func ResolveListPagination(limitStr, offsetStr, pageStr string) (limit int, offset int, err error) {
	if limitStr == "" {
		limitStr = "10"
	}
	limit, err = strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		return 0, 0, errors.New("limit must be a positive integer")
	}

	if pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return 0, 0, errors.New("page must be a positive integer")
		}
		return limit, (page - 1) * limit, nil
	}

	if offsetStr == "" {
		offsetStr = "0"
	}
	offset, err = strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return 0, 0, errors.New("offset must be a non-negative integer")
	}
	return limit, offset, nil
}

type listPagination struct {
	Limit int `json:"limit"`
	// TotalPages is exposed as page_size: how many pages exist for this total/limit (ceil(total/limit)).
	TotalPages int `json:"page_size"`
	Page       int `json:"page"`
	Offset     int `json:"offset"`
	Total      int `json:"total"`
}

func paginationTotalPages(total, limit int) int {
	if limit < 1 || total == 0 {
		return 0
	}
	return (total + limit - 1) / limit
}

func paginationCurrentPage(offset, limit int) int {
	if limit < 1 {
		return 1
	}
	return offset/limit + 1
}

type trainerListData struct {
	Trainers   []repository.Trainer `json:"trainers"`
	Pagination listPagination     `json:"pagination"`
}

type fighterListData struct {
	Fighters   []repository.Fighter `json:"fighters"`
	Pagination listPagination     `json:"pagination"`
}
