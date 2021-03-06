package email

import (
	"fmt"
	"net/url"

	mailgun "gopkg.in/mailgun/mailgun-go.v1"
)

const (
	resetBaseURL   = "http://localhost:3000/password/reset"
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

	resetSubject      = "Fakeoku Account Recovery"
	resetTextTemplate = `Hi There!
	
	It appears you have requested a password reset. If this was you,
	please follow the link below to update your password:
	
	%s
	
	If you are asked for a token please use the following value:
	
	%s
	
	If you did not request a password reset, you can ignore this email
	and your account will not be updated.
	
	Best,
	The DevOps Team`
	resetHTMLTemplate = `Hi There!<br/>
	<br/>
	It appears you have requested a password reset. If this was you,
	please follow the link below to update your password:<br/>
	<br/>
	<a href="%s">%s</a><br/>
	<br/>
	If you are asked for a token please use the following value:<br/>
	<br/>
	%s<br/>
	<br/>
	If you did not request a password reset, you can ignore this email
	and your account will not be updated.<br/>
	<br/>
	Best,<br/>
	The DevOps Team`
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

func (client *Client) SendResetPasswordMessage(recipientName string, recipientEmail string, token string) error {
	values := url.Values{}
	values.Set("token", token)
	resetURL := resetBaseURL + "?" + values.Encode()
	resetText := fmt.Sprintf(resetTextTemplate, resetURL, token)
	resetHTML := fmt.Sprintf(resetHTMLTemplate, resetURL, resetURL, token)
	return client.SendHTMLMessage(client.from, buildEmail(recipientName, recipientEmail), resetSubject, resetText, resetHTML)
}

func buildEmail(name, email string) string {
	if name == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}
