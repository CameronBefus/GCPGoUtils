package email

import (
	"net/http"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const envSendGridAPI = `SENDGRID_API_KEY`
const envTestKey = `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`
const envFromEmailKey = `FROM_EMAIL`
const envFromUserKey = `FROM_USER`

func init() {
	viper.SetDefault(envSendGridAPI, envTestKey)
	viper.SetDefault(envFromEmailKey, `no-reply@mjdevelopment`)
	viper.SetDefault(envFromUserKey, `Chief Tester`)
}

// MailInit - creates basic email object, which can then be further customized if necessary
// dest string is a comma delimited string of one or more email addresses
func MailInit(subject string, dest string, plain string, html string) *mail.SGMailV3 {
	m := new(mail.SGMailV3)
	m.SetFrom(mail.NewEmail(viper.GetString(envFromUserKey), viper.GetString(envFromEmailKey)))
	m.Subject = subject
	p := mail.NewPersonalization()

	a := strings.Split(dest, ",")
	for _, to := range a {
		p.AddTos(mail.NewEmail("", to))
	}
	m.AddPersonalizations(p)

	if len(plain) > 0 {
		m.AddContent(mail.NewContent("text/plain", plain))
	}
	if len(html) > 0 {
		m.AddContent(mail.NewContent("text/html", html))
	}
	return m
}

// Send - launches the prepared email
func Send(m *mail.SGMailV3) bool {
	client := sendgrid.NewSendClient(viper.GetString(envSendGridAPI))
	response, err := client.Send(m)
	if err != nil {
		log.Error(err)
		return false
	}
	if response.StatusCode != http.StatusAccepted {
		log.Warning(response)
		return false
	}
	return true
}
