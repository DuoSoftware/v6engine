package repositories

import ()

func Create(code string) AbstractETL {

	var ETL AbstractETL
	switch code {
	case "KIBANA":
		ETL = KibanaRepository{}
	case "POSTGRESv5":
		ETL = Postgres5Repository{}
	}
	return ETL
}
