// using SendGrid's Go Library
// https://github.com/sendgrid/sendgrid-go
package email

import (
	"testing"
)

const testEmail = `wlsSupport@mjlogs.com,mike.adams@mjlogs.com,cameron@befus.net,mjlogswls@gmail.com`

func Test_sendSimpleEmail(t *testing.T) {

	m := MailInit(`test msg`, testEmail, "this is a simple email with no html", "")
	if !Send(m) {
		t.Error()
	}
	msg := `
	<!DOCTYPE html>
	<html>
	<body>
	<h1>Heading 1</h1>
	<h2>Heading 2</h2>
	<h3>Heading 3</h3>
	<p>Here is a paragraph</p>
	</body>
	</html>
	`

	m2 := MailInit(`test msg2`, testEmail, "alt", msg)
	if !Send(m2) {
		t.Error()
	}

}
