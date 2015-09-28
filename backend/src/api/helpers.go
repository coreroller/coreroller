package api

const (
	defaultPage    uint64 = 1
	defaultPerPage uint64 = 100
)

// validatePaginationParams validates the pagination parameters provided,
// setting them to the default values in case they are invalid.
func validatePaginationParams(page, perPage uint64) (uint64, uint64) {
	if page < 1 {
		page = defaultPage
	}

	if perPage < 1 {
		perPage = defaultPerPage
	}

	return page, perPage
}
