// Package mailer contains logic related to sending emails.
package mailer

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	ht "html/template"
	"path/filepath"
	tt "text/template"
	"time"

	"github.com/mailtrap/mailtrap-go"
	"github.com/wneessen/go-mail"
)

//go:embed "templates"
var templateFS embed.FS

// Mailer has everything to interact with an SMTP server.
type Mailer struct {
	client  *mailtrap.Client
	client2 *mail.Client // used to connect to an SMTP server
	sender  string       // sender's info, like "Alice Smith <alice@example.com>"
}

func New(apiToken, sender string) (*Mailer, error) {
	cli, err := mailtrap.NewClient(apiToken)
	if err != nil {
		return nil, fmt.Errorf("mailer: New: %w", err)
	}
	m := &Mailer{
		client: cli,
		sender: sender,
	}

	return m, nil
}

func New2(host string, port int, username, password, sender string) (*Mailer, error) {
	cli, err := mail.NewClient(
		host,
		mail.WithSMTPAuth(mail.SMTPAuthLogin), // to choose login as SMTP server's authentication method
		mail.WithPort(port),                   // port client uses to connect to SMTP server.
		mail.WithUsername(username),
		mail.WithPassword(password),
		mail.WithTimeout(5*time.Second), // client hangs up after this duration.
	)
	if err != nil {
		return nil, err
	}
	mailer := &Mailer{
		client2: cli,
		sender:  sender,
	}

	return mailer, nil
}

func (m *Mailer) Client() *mailtrap.Client {
	return m.client
}

// Send renders content of templateFile with data, and sends the result as email to recipient address.
func (m *Mailer) Send(recipient, templateFile string, data any) error {
	// create a *text/template.Template value to parse and contain templates's content defined in templateFile from an embed.FS.
	txtTmpl, err := tt.New("").ParseFS(templateFS, filepath.Join("templates", templateFile))
	if err != nil {
		return err
	}
	subject := new(bytes.Buffer)
	if err = txtTmpl.ExecuteTemplate(subject, "subject", data); err != nil { // write the parsed content of a defined template to a buffer
		return err
	}
	plainBody := new(bytes.Buffer)
	if err = txtTmpl.ExecuteTemplate(plainBody, "plainBody", data); err != nil {
		return err
	}

	// create a *html/template.Template value to parse and contain templates's content defined in templateFile from an embed.FS.
	htmlTmpl, err := ht.New("").ParseFS(templateFS, filepath.Join("templates", templateFile))
	if err != nil {
		return err
	}
	htmlBody := new(bytes.Buffer)
	if err = htmlTmpl.ExecuteTemplate(htmlBody, "htmlBody", data); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const retry = 3
	for i := range retry {
		if _, _, err = m.client.Send(ctx, &mailtrap.SendRequest{
			From:    mailtrap.Address{Email: m.sender, Name: "MoviesDB"},
			To:      []mailtrap.Address{{Email: recipient}},
			Subject: subject.String(),
			Text:    plainBody.String(),
			HTML:    htmlBody.String(),
		}); err != nil {
			if i != retry-1 {
				time.Sleep(time.Second)
			}
		}
	}

	return err
}

// Send2 renders content of templateFile with data, and sends the result as email to recipient address.
func (m *Mailer) Send2(recipient, templateFile string, data any) error {
	// create a *text/template.Template value to parse and contain templates's content defined in templateFile from an embed.FS.
	txtTmpl, err := tt.New("").ParseFS(templateFS, filepath.Join("templates", templateFile))
	if err != nil {
		return err
	}
	subject := new(bytes.Buffer)
	if err = txtTmpl.ExecuteTemplate(subject, "subject", data); err != nil { // write the parsed content of a defined template to a buffer
		return err
	}
	plainBody := new(bytes.Buffer)
	if err = txtTmpl.ExecuteTemplate(plainBody, "plainBody", data); err != nil {
		return err
	}

	// create a *html/template.Template value to parse and contain templates's content defined in templateFile from an embed.FS.
	htmlTmpl, err := ht.New("").ParseFS(templateFS, filepath.Join("templates", templateFile))
	if err != nil {
		return err
	}
	htmlBody := new(bytes.Buffer)
	if err = htmlTmpl.ExecuteTemplate(htmlBody, "htmlBody", data); err != nil {
		return err
	}

	msg := mail.NewMsg() // the whole content of email.
	if err = msg.To(recipient); err != nil {
		return err
	}
	if err = msg.From(m.sender); err != nil {
		return err
	}
	msg.Subject(subject.String())
	msg.SetBodyString(mail.TypeTextPlain, plainBody.String())      // set the content of the email's body; this case, content is plain text.
	msg.AddAlternativeString(mail.TypeTextHTML, htmlBody.String()) // set the content of the email's alternative body; this case, content is html.

	const retry = 10
	for i := range retry {
		if err = m.client2.DialAndSend(msg); err != nil {
			if i != retry-1 {
				time.Sleep(time.Second)
			}
		}
	}

	return err
}
