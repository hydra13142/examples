package main

import (
	"fmt"
	"net/smtp"
	"strings"
)

/*
 *	user : example@example.com login smtp server user
 *	password: xxxxx login smtp server password
 *	host: smtp.example.com:port   smtp.163.com:25
 *	to: example@example.com;example1@163.com;example2@sina.com.cn;...
 *  subject:The subject of mail
 *  body: The content of mail
 *  mailtyoe: mail type html or text
 */

func SendMail(user, password, host, to, subject, body, mailtype string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])
	var Content_Type string
	if mailtype == "html" {
		Content_Type = "Content-Type: text/html; charset=UTF-8"
	} else {
		Content_Type = "Content-Type: text/plain; charset=UTF-8"
	}

	msg := []byte("To: " + to + "\r\n" +
		"From: " + user + "<" + user + ">\r\n" +
		"Subject: " + subject + "\r\n" +
		Content_Type + "\r\n\r\n" +
		body)
	send_to := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, user, send_to, msg)
	return err
}

func main() {
	user := "snake_dragon@126.com"
	password := "6ys97pt"
	host := "smtp.126.com:25"
	to := "1051474363@qq.com;ironandconceret@sina.com.cn"

	subject := "Test send email by golang"

	body := `
	<html>
	<body>
	<h3>
	"Test send email by golang"
	</h3>
	</body>
	</html>
	`
	fmt.Println("send email")
	err := SendMail(user, password, host, to, subject, body, "html")
	if err != nil {
		fmt.Println("send mail error!")
		fmt.Println(err)
	} else {
		fmt.Println("send mail success!")
	}

}
