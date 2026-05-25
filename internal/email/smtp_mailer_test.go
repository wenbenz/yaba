package email_test

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"yaba/internal/email"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// smtpLineResult is the action to take after processing one SMTP client line.
type smtpLineResult int

const (
	smtpContinue  smtpLineResult = iota // keep reading
	smtpDataReady                       // entering DATA mode
	smtpBodyDone                        // DATA body complete — body is ready
	smtpQuit                            // QUIT received — close connection
)

// handleSMTPLine processes one line from the SMTP client, writes the
// appropriate response to conn, and returns the next action to take.
func handleSMTPLine(conn net.Conn, body *strings.Builder, line string, inData bool) smtpLineResult {
	switch {
	case strings.HasPrefix(line, "EHLO"):
		_, _ = fmt.Fprintln(conn, "250-test\r\n250 AUTH PLAIN LOGIN")
	case strings.HasPrefix(line, "AUTH"):
		_, _ = fmt.Fprintln(conn, "235 OK")
	case strings.HasPrefix(line, "MAIL FROM"), strings.HasPrefix(line, "RCPT TO"):
		_, _ = fmt.Fprintln(conn, "250 OK")
	case line == "DATA":
		_, _ = fmt.Fprintln(conn, "354 Start input")

		return smtpDataReady
	case inData && line == ".":
		_, _ = fmt.Fprintln(conn, "250 OK")

		return smtpBodyDone
	case inData:
		body.WriteString(line + "\n")
	case strings.HasPrefix(line, "QUIT"):
		_, _ = fmt.Fprintln(conn, "221 Bye")

		return smtpQuit
	}

	return smtpContinue
}

// runFakeSMTPServer starts a minimal fake SMTP server on ln and returns a
// channel that receives the DATA body once a complete message is accepted.
func runFakeSMTPServer(t *testing.T, ln net.Listener) <-chan string {
	t.Helper()

	captured := make(chan string, 1)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}

		defer conn.Close()

		var body strings.Builder

		scanner := bufio.NewScanner(conn)
		inData := false

		_, _ = fmt.Fprintln(conn, "220 test ready")

		for scanner.Scan() {
			line := scanner.Text()

			switch handleSMTPLine(conn, &body, line, inData) {
			case smtpDataReady:
				inData = true
			case smtpBodyDone:
				captured <- body.String()

				inData = false
			case smtpQuit:
				return
			case smtpContinue:
			}
		}
	}()

	return captured
}

func TestSMTPMailerSend(t *testing.T) {
	t.Parallel()

	lc := &net.ListenConfig{}

	ln, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	defer ln.Close()

	_, port, _ := net.SplitHostPort(ln.Addr().String())

	captured := runFakeSMTPServer(t, ln)

	m := email.NewSMTPMailer(email.SMTPConfig{
		Host:     "127.0.0.1",
		Port:     port,
		Username: "user",
		Password: "pass",
		From:     "from@example.com",
	})

	err = m.Send("to@example.com", "Test Subject", "<p>Hello world</p>")
	require.NoError(t, err)

	body := <-captured
	assert.Contains(t, body, "Subject: Test Subject")
	assert.Contains(t, body, "From: from@example.com")
	assert.Contains(t, body, "To: to@example.com")
	assert.Contains(t, body, "<p>Hello world</p>")
}

func TestNoopMailer(t *testing.T) {
	t.Parallel()

	var m email.Mailer = email.NoopMailer{}

	assert.NoError(t, m.Send("to@example.com", "subject", "<p>body</p>"))
}
