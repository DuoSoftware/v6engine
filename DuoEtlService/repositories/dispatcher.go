package repositories

import (
	"duov6.com/DuoEtlService/logger"
	"duov6.com/DuoEtlService/messaging"
)

func Dispatch(request *messaging.ETLRequest) {
	logger.Log("Initializing Repository Dispatcher....")
	repos := getRepositories(request)
	response := startAtomicOperation(request, repos)

	if response.IsSuccess {
		logger.Log("Dispatch Successful!")
	} else {
		logger.Log("Dispatch Failed!")
	}
	logger.Log(response.Message)
	logger.Log("\n")
	logger.Log("\n")

}

func getRepositories(request *messaging.ETLRequest) []AbstractETL {
	var outRepos []AbstractETL
	outRepos = make([]AbstractETL, len(request.Configuration.EtlConfig))
	index := 0

	for configName, _ := range request.Configuration.EtlConfig {
		repo := Create(configName)
		outRepos[index] = repo
		index++
	}

	return outRepos
}

func startAtomicOperation(request *messaging.ETLRequest, repositoryList []AbstractETL) (response messaging.ETLResponse) {

	var states []bool
	var msgs []string
	states = make([]bool, len(repositoryList))
	msgs = make([]string, len(repositoryList))
	ifErrorMsg := ""
	index := 0

	for _, repository := range repositoryList {
		if repository != nil {
			logger.Log("Executing repository : " + repository.GetETLName())

			tmpResponse := repository.ExecuteETLService(request)
			states[index] = tmpResponse.IsSuccess
			msgs[index] = tmpResponse.Message

			response = tmpResponse
			if tmpResponse.IsSuccess {
				logger.Log("Executing repository : " + repository.GetETLName() + " - Success")
			} else {
				logger.Log("Executing repository : " + repository.GetETLName() + " - Failed")
			}

			index++
		} else {
			logger.Log("NIL REPOSITORY FOUND!")
			continue
		}
	}

	for key, value := range states {
		if value == false {
			ifErrorMsg = msgs[key]
			break
		}
	}

	if ifErrorMsg != "" {
		response.IsSuccess = false
		response.Message = "Some error occured in one of repositories : " + ifErrorMsg
	} else {
		response.IsSuccess = true
		response.Message = "All repositories executed Successfully!"
	}

	return response
}
