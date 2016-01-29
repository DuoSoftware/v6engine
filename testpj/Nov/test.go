package main

import (
	"duov6.com/queryparser"
	"fmt"
)

func main() {
	//fmt.Println(queryparser.GetDataStoreQuery("SELECT name, Id, age from Student s2 , Game g2 where age >= 10 and course = 'SLIIT' order by Id ASC;"))
	//fmt.Println(queryparser.GetElasticQuery("SELECT tttt, Id, age from Student s1, game g2 where ddf = 5@00 AND kappa not between 50 AND 100 AND kat not in ('asdf', 'wert') OR Country LIKE '%land%' order by ppds, ggg desc;", "com.duoworld.com", "test"))
	//fmt.Println(queryparser.GetElasticQuery("SELECT tttt, Id, age from Student s1, game g2;", "com.duoworld.com", "test"))
	//fmt.Println(queryparser.GetDataStoreQuery("SELECT tttt, Id, age from Student s1, game g2 order by age desc, name ;", "com.duoworld.com", "test"))
	//fmt.Println(queryparser.GetElasticQuery("select * from chatmessages where (too='usr1' AND frm='usr2') OR (too= 'usr2' AND frm='usr1');", "com.duoworld.com", "test"))

	fmt.Println(queryparser.GetDataStoreQuery("SELECT tttt, Id, age from Student s1, game g2 where value = '50 0' AND kappa not between 50 AND 100 AND kat not in ('asdf', 'wert') order by ppds, ggg desc;", "com.duoworld.com", "test"))

}
