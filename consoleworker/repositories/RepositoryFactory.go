package repositories

func Create(code string) AbstractRepository {
	var repository AbstractRepository
	switch code {
	case "BulkProcessor":
		repository = BulkProcessor{}
	case "SmoothFlow":
		repository = SmoothFlowProcessor{}
	default:
		repository = NullRepository{}
	}

	return repository
}
