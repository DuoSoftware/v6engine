package repositories

func Create(code string) AbstractRepository {
	var repository AbstractRepository
	switch code {
	case "COUCH":
		repository = CouchRepository{}
	case "ELASTIC":
		repository = ElasticRepository{}
	case "REDIS":
		repository = RedisRepository{}
	case "MONGO":
		repository = MongoRepository{}
	case "CASSANDRA":
		repository = CassandraRepository{}
	case "HIVE":
		repository = HiveRepository{}
	case "POSTGRES":
		repository = PostgresRepository{}
	case "GoogleDataStore":
		repository = GoogleDataStoreRepository{}
	case "GoogleBigTable":
		repository = GoogleBigTableRepository{}
	case "MSSQL":
		repository = MssqlRepository{}
	case "CLOUDSQL":
		repository = CloudSqlRepository{}
	}

	return repository
}
