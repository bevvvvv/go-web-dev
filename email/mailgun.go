package email

import (
	"fmt"

	mailgun "gopkg.in/mailgun/mailgun-go.v1"
)

const (
	welcomeSubject = "Welcome to Fakeoku!"
	welcomeText    = `Hi There!
	
	Welcome to the Fakeoku site! I hope you enjoy using the application.
	
	Best,
	The DevOps Team`
	welcomeHTML = `Hi There!<br/>
	<br/>
	Welcome to the
	<a href="http://localhost:3000">Fakeoku</a> site! I hope you enjoy using
	the application.<br/>
	<br/>
	Best,<br/>
	The DevOps Team
	`
)

type Client struct {
	from          string
	mailgunClient mailgun.Mailgun
}

type ClientConfig func(*Client)

func WithMailgun(apiKey string, publicAPIKey string, domain string) ClientConfig {
	return func(client *Client) {
		client.mailgunClient = mailgun.NewMailgun(domain, apiKey, publicAPIKey)
	}
}

func WithSender(name, email string) ClientConfig {
	return func(client *Client) {
		client.from = buildEmail(name, email)
	}
}

func NewClient(configs ...ClientConfig) *Client {
	client := Client{
		from: "support@devops.gg",
	}
	for _, config := range configs {
		config(&client)
	}
	return &client
}

func (client *Client) SendMessage(sender string, recipient string, subject string, text string) error {
	message := mailgun.NewMessage(sender, subject, text, recipient)
	_, _, err := client.mailgunClient.Send(message)
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) SendHTMLMessage(sender string, recipient string, subject string, text string, html string) error {
	message := mailgun.NewMessage(sender, subject, text, recipient)
	message.SetHtml(html)
	_, _, err := client.mailgunClient.Send(message)
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) SendWelcomeMessage(recipientName string, recipientEmail string) error {
	return client.SendHTMLMessage(client.from, buildEmail(recipientName, recipientEmail), welcomeSubject, welcomeText, welcomeHTML)
}

func buildEmail(name, email string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}
