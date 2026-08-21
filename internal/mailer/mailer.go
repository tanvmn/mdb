// Package mailer contains logic related to sending emails.
package mailer

import (
	"bytes"
	"embed"
	ht "html/template"
	"path/filepath"
	tt "text/template"
	"time"

	"github.com/wneessen/go-mail"
)

//go:embed "templates"
var templateFS embed.FS

// Mailer has everything to interact with an SMTP server.
type Mailer struct {
	client *mail.Client // used to connect to an SMTP server
	sender string       // sender's info, like "Alice Smith <alice@example.com>"
}

func New(host string, port int, username, password, sender string) (*Mailer, error) {
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
		client: cli,
		sender: sender,
	}

	return mailer, nil
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

	for i := range 3 {
		if err = m.client.DialAndSend(msg); err != nil {
			if i != 2 {
				time.Sleep(time.Second)
			}
		}
	}

	return err
}
