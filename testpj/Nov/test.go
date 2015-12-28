package main

import (
	"duov6.com/queryparser"
	"fmt"
)

func main() {
	//fmt.Println(queryparser.GetDataStoreQuery("SELECT name, Id, age from Student s2 , Game g2 where age >= 10 and course = 'SLIIT' order by Id ASC;"))
	fmt.Println(queryparser.GetCloudSQLQuery("SELECT tttt, Id, age from Student s1, game g2 where table = 500 AND value not between 50 AND 100 AND kat not in ('asdf', 'wert');", "com.duoworld.com", "test"))
	//fmt.Println(queryparser.GetCloudSQLQuery("SELECT tttt, Id, age from Student s1, game g2;", "com.duoworld.com", "test"))
	//fmt.Println(queryparser.GetCloudSQLQuery("SELECT tttt, Id, age from Student s1, game g2 order by age desc, name ;", "com.duoworld.com", "test"))
}
