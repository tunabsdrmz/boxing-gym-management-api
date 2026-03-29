package handler

import (
	"net/url"
	"strconv"

	"github.com/tunabsdrmz/boxing-gym-management/internal/repository"
)

func fighterListRequest(q url.Values, forceLimit, forceOffset *int) (repository.GetAllFightersRequest, int, int, error) {
	var limit, offset int
	var err error
	if forceLimit != nil && forceOffset != nil {
		limit, offset = *forceLimit, *forceOffset
	} else {
		limit, offset, err = ResolveListPagination(q.Get("limit"), q.Get("offset"), q.Get("page"))
		if err != nil {
			return repository.GetAllFightersRequest{}, 0, 0, err
		}
	}
	sort := q.Get("sort")
	order := q.Get("order")
	sortAsc := order == "asc"
	return repository.GetAllFightersRequest{
		Limit:         strconv.Itoa(limit),
		Offset:        strconv.Itoa(offset),
		Search:        q.Get("q"),
		WeightClass:   q.Get("weight_class"),
		FighterStatus: q.Get("fighter_status"),
		SortField:     sort,
		SortAsc:       sortAsc,
	}, limit, offset, nil
}
