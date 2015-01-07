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
	}
	return repository
}
