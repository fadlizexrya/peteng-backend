package mailer

import (
	"crypto/tls"
	"fmt"
	"html"
	"log"
	"mime"
	"net"
	"net/smtp"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Mailer mengirim email verifikasi lewat SMTP (Gmail/Resend/penyedia lain).
// Kalau SMTP tidak dikonfigurasi, aplikasi tetap jalan dalam mode dev dan
// kode verifikasi hanya ditulis ke log backend.
type Mailer struct {
	Host      string
	Port      int
	Username  string
	Password  string
	From      string
	FromName  string
	AppName   string
	VerifyURL string
}

func NewFromEnv() *Mailer {
	port := 587

	if value := strings.TrimSpace(os.Getenv("SMTP_PORT")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			port = parsed
		}
	}

	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))

	if from == "" {
		from = strings.TrimSpace(os.Getenv("SMTP_USERNAME"))
	}

	appName := strings.TrimSpace(os.Getenv("APP_NAME"))

	if appName == "" {
		appName = "PETENG"
	}

	verifyURL := strings.TrimSpace(os.Getenv("APP_VERIFY_URL"))

	if verifyURL == "" {
		verifyURL = "http://127.0.0.1:8080/api/warga/verify-email"
	}

	return &Mailer{
		Host:      strings.TrimSpace(os.Getenv("SMTP_HOST")),
		Port:      port,
		Username:  strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		Password:  os.Getenv("SMTP_PASSWORD"),
		From:      from,
		FromName:  strings.TrimSpace(os.Getenv("SMTP_FROM_NAME")),
		AppName:   appName,
		VerifyURL: verifyURL,
	}
}

func (m *Mailer) Enabled() bool {
	return m != nil && m.Host != "" && m.From != ""
}

func (m *Mailer) SendVerification(to string, name string, code string) error {
	if !m.Enabled() {
		return fmt.Errorf("SMTP belum dikonfigurasi")
	}

	if strings.TrimSpace(m.FromName) == "" {
		m.FromName = m.AppName
	}

	separator := "?"

	if strings.Contains(m.VerifyURL, "?") {
		separator = "&"
	}

	verifyLink := m.VerifyURL + separator + "email=" + url.QueryEscape(to) + "&code=" + url.QueryEscape(code)

	subject := fmt.Sprintf("Kode verifikasi %s: %s", m.AppName, code)

	body := buildVerificationHTML(m.AppName, name, code, verifyLink)

	message := strings.Join([]string{
		"From: " + mime.QEncoding.Encode("UTF-8", m.FromName) + " <" + m.From + ">",
		"To: <" + to + ">",
		"Subject: " + mime.QEncoding.Encode("UTF-8", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"UTF-8\"",
		"Content-Transfer-Encoding: 8bit",
		"Date: " + time.Now().Format(time.RFC1123Z),
		"",
		body,
	}, "\r\n")

	address := net.JoinHostPort(m.Host, strconv.Itoa(m.Port))

	var client *smtp.Client

	if m.Port == 465 {
		connection, err := tls.DialWithDialer(
			&net.Dialer{Timeout: 20 * time.Second},
			"tcp",
			address,
			&tls.Config{ServerName: m.Host},
		)
		if err != nil {
			return err
		}

		client, err = smtp.NewClient(connection, m.Host)
		if err != nil {
			connection.Close()
			return err
		}
	} else {
		connection, err := net.DialTimeout("tcp", address, 20*time.Second)
		if err != nil {
			return err
		}

		client, err = smtp.NewClient(connection, m.Host)
		if err != nil {
			connection.Close()
			return err
		}

		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: m.Host}); err != nil {
				client.Close()
				return err
			}
		}
	}

	defer client.Close()

	if m.Username != "" {
		auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)

		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(m.From); err != nil {
		return err
	}

	if err := client.Rcpt(to); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}

	if _, err := writer.Write([]byte(message)); err != nil {
		writer.Close()
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	if err := client.Quit(); err != nil {
		log.Printf("SMTP Quit: %v", err)
	}

	return nil
}

func buildVerificationHTML(appName string, name string, code string, verifyLink string) string {
	displayName := strings.TrimSpace(name)

	if displayName == "" {
		displayName = "Warga"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="id">
<body style="margin:0;padding:0;background:#F1F1EF;font-family:Segoe UI,Arial,sans-serif;">
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background:#F1F1EF;padding:32px 12px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="max-width:520px;background:#ffffff;border-radius:18px;overflow:hidden;">
          <tr>
            <td style="background:#172554;padding:26px 30px;">
              <div style="color:#ffffff;font-size:22px;font-weight:800;letter-spacing:1px;">%s</div>
              <div style="color:#E8D98A;font-size:12px;margin-top:4px;">Petakan Gelap, Temukan Terang</div>
            </td>
          </tr>
          <tr>
            <td style="padding:30px;">
              <p style="margin:0 0 8px;color:#172554;font-size:16px;font-weight:700;">Halo, %s</p>
              <p style="margin:0 0 22px;color:#5B6070;font-size:14px;line-height:1.6;">
                Terima kasih sudah mendaftar di %s. Masukkan kode berikut di aplikasi
                untuk memverifikasi email kamu.
              </p>
              <div style="background:#F1F1EF;border:1px dashed #172554;border-radius:12px;padding:18px;text-align:center;">
                <div style="color:#5B6070;font-size:11px;letter-spacing:1.5px;text-transform:uppercase;">Kode verifikasi</div>
                <div style="color:#172554;font-size:34px;font-weight:800;letter-spacing:10px;margin-top:6px;">%s</div>
              </div>
              <p style="margin:22px 0 14px;color:#5B6070;font-size:13px;line-height:1.6;">
                Kode berlaku 15 menit. Kalau kamu tidak merasa mendaftar, abaikan email ini.
              </p>
              <a href="%s" style="display:inline-block;background:#172554;color:#ffffff;text-decoration:none;font-size:14px;font-weight:600;padding:12px 22px;border-radius:10px;">
                Verifikasi lewat link
              </a>
              <p style="margin:22px 0 0;color:#9AA0AE;font-size:11px;line-height:1.5;word-break:break-all;">%s</p>
            </td>
          </tr>
          <tr>
            <td style="background:#F8F8F6;padding:16px 30px;color:#9AA0AE;font-size:11px;">
              Email otomatis dari %s. Jangan balas email ini.
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`,
		html.EscapeString(appName),
		html.EscapeString(displayName),
		html.EscapeString(appName),
		html.EscapeString(code),
		html.EscapeString(verifyLink),
		html.EscapeString(verifyLink),
		html.EscapeString(appName),
	)
}
