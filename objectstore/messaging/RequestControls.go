package messaging

type RequestControls struct {
	SecurityToken string
	SendMetaData  string
	Namespace     string
	Class         string
	Operation     string //CREATE, READ, UPDATE, DELETE, SPECIAL
	Multiplicity  string //SINGLE, MULTIPLE

	Id         string //Doesn't apply for GET
	NewVersion string
}
