package email

// Mailer sends transactional emails.
type Mailer interface {
	Send(to, subject, htmlBody string) error
}

// NoopMailer discards all mail silently. Used when SMTP is not configured.
type NoopMailer struct{}

func (NoopMailer) Send(_, _, _ string) error { return nil }
