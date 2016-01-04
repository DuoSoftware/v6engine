package repositories

func Create(code string) AbstractRepository {
	var repository AbstractRepository
	switch code {
	case "ES":
		repository = ElasticSearch{}
	case "MYSQL":
		repository = CommonSQL{}
	case "MSSQL":
		repository = CommonSQL{}
	case "CSQL":
		repository = CommonSQL{}
	case "PSQL":
		repository = CommonSQL{}
	case "POSTGRES":
		repository = CommonSQL{}
	case "HSQL":
		repository = CommonSQL{}
	case "CDS":
		repository = GoogleCloudDataStore{}
	}

	return repository
}
