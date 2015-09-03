package agentCore


type AgentLogger struct {
	logChannel       chan LogLine
	isChannelCreated bool
}

type LogLine struct {
	Output string
	//MType  int
}

func (l AgentLogger) Log(Lable string, mType int) {

	line := LogLine{Output: Lable} //, MType: mType}

	if !l.isChannelCreated {
		l.isChannelCreated = true
		l.logChannel = make(chan LogLine)
		go consumeLogLines(l)
	}

	l.logChannel <- line

}

func consumeLogLines(l AgentLogger) {
	client := GetInstance().Client

	select {
	case i := <-l.logChannel:
		if (client.ListenerName != ""){
			client.ClientCommand(client.ListenerName, "log", "output", i)	
		}
	}
}
