package main

import (
	"duov6.com/queryparser"
	"fmt"
)

func main() {
	//fmt.Println(queryparser.GetDataStoreQuery("SELECT name, Id, age from Student s2 , Game g2 where age >= 10 and course = 'SLIIT' order by Id ASC;"))
	fmt.Println(queryparser.GetCloudSQLQuery("SELECT tttt, Id, age from Student s1, game g2 where asdf = 500 AND kappa not between 50 AND 100 AND kat in ('asdf', 'wert') order by ppds;", "com.duoworld.com", "test"))
	//fmt.Println(queryparser.GetCloudSQLQuery("SELECT tttt, Id, age from Student s1, game g2;", "com.duoworld.com", "test"))
	//fmt.Println(queryparser.GetCloudSQLQuery("SELECT tttt, Id, age from Student s1, game g2 order by age desc, name ;", "com.duoworld.com", "test"))
}
