package main

import (
	"duov6.com/queryparser/lexer"
	"duov6.com/queryparser/messaging"
	"duov6.com/queryparser/truncator"
	"fmt"
)

type queryparser struct {
	Request *messaging.ParserRequest
}

func main() {

	var parser queryparser
	parser.ProcessQuery()
}

func (q *queryparser) ProcessQuery() {

	q.Request = &messaging.ParserRequest{}
	q.Request.Query = "SELECT name, Id, age from Student s2 , Game g2 where age >= 10 & course = 'SLIIT' order by Id ASC"
	//q.Request.Query = "SELECT name, Id, age from Student s1, Game g2"
	q.Request.Body = make(map[string]string)

	//Get Normalized Tokens
	lex := lexer.Tokenizer{}
	response := lex.GetTokens(q.Request)
	fmt.Println(response)

	q.Request.Body = response.Body

	//Extract Query Attributes

	tru := truncator.Slicer{}
	response2 := tru.Begin(q.Request)
	fmt.Println("-------------")
	fmt.Println(response2)

}
