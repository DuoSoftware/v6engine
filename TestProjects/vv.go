package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	dd := `{"Body":"<html><head><title>Digin - Registration Confirmation</title></head><title >Digin-Registration Confirmation</title><body><h2>Digin Registration - Confirmation Mail</h2><h3 class=\"\">Dear @@name@@,</h3><p>We wanted to follow up on your DigIn trial - Thanks for signing up! We hope over the coming days you'll see how easy it is to gain actionable insight from almost any data source with our solution. We am on hand should you have any questions whatsoever as you get to grips with DigIn, please don't hesitate to call or email at any point.</p><p>You can also check out our YouTube videos, or the support documents on the website:</p><p>DigIn - YouTube<br/>DigIn - Support<br/>DigIn - Showcase</p> <p>DataSet is successfully created with <b>@@dataSet@@</b>, Upload the data with the mentioned ETL tool with the following credentials and start Diging.</p><i>Clientid :302371564513-dqpdjhgds4pejr1tpi1ke735m894m3lg.apps.googleusercontent.com<br/>ClientSecret : IArNi-rtAwh0GkhytzBqtTR-<br/>Projectid : thematic-scope-112013<br/>BucketName : diginbeta<br/> </i><br/> <p>We will check back in the coming days to see if you have any questions for me.</p><p>Many thanks,</p><h3>DigIn Team</h3></body></html>","Owner":"digin","TemplateID":"registration_confirmation2","Title":"Account Verification2"}`
	ll := make(map[string]interface{})
	err := json.Unmarshal([]byte(dd), &ll)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(ll)
	}
}
