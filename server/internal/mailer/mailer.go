package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/arinsuda/movie-hub/config"
)

type Mailer struct {
	apiKey      string
	senderEmail string
	senderName  string
	enabled     bool
	devLog      bool
	devOut      io.Writer
	httpClient  *http.Client
}

// PrintDev writes a DEV_MAIL_LOG banner. It is the only place codes and links are printed.
func PrintDev(w io.Writer, to, kind string, lines ...string) {
	fmt.Fprintf(w, "\n========== [DEV MAIL] ==========\n[DEV MAIL] to=%s kind=%s\n", to, kind)
	for _, l := range lines {
		fmt.Fprintf(w, "  %s\n", l)
	}
	fmt.Fprint(w, "================================\n\n")
}

func New(cfg config.BrevoConfig) *Mailer {
	return &Mailer{
		apiKey:      cfg.APIKey,
		senderEmail: cfg.SenderEmail,
		senderName:  cfg.SenderName,
		enabled:     cfg.Enabled || cfg.DevLog,
		devLog:      cfg.DevLog,
		devOut:      os.Stdout,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *Mailer) IsEnabled() bool {
	return m.enabled
}

func (m *Mailer) SendVerificationEmail(toEmail, username, verifyURL string) error {
	subject := "ยืนยันอีเมลของคุณ — POPCORN"
	if m.devLog {
		PrintDev(m.devOut, toEmail, "verify", "link: "+verifyURL)
		return nil
	}
	body := buildVerifyEmailBody(username, verifyURL)
	return m.send(toEmail, username, subject, body)
}

// SendLinkCodeEmail sends a 6-digit code used to confirm linking a PLANZI account.
func (m *Mailer) SendLinkCodeEmail(toEmail, username, code string) error {
	if m.devLog {
		PrintDev(m.devOut, toEmail, "link_code", "code: "+code)
		return nil
	}
	subject := "รหัสยืนยันการเชื่อมบัญชี — POPCORN"
	body := fmt.Sprintf(`<!DOCTYPE html><html><body style="margin:0;padding:32px 16px;background:#050505;font-family:'Helvetica Neue',Arial,sans-serif;">
<div style="max-width:440px;margin:0 auto;background:#0f0f0f;border:1px solid #1c1c1c;border-radius:4px;padding:40px;text-align:center;">
<div style="font-size:18px;font-weight:700;letter-spacing:4px;color:#fff;">POPCORN</div>
<p style="color:#fff;font-size:16px;margin:24px 0 8px;">สวัสดี %s</p>
<p style="color:#8f8f8f;font-size:14px;line-height:1.7;">ใช้รหัสนี้เพื่อยืนยันการเชื่อมบัญชี PLANZI กับ POPCORN</p>
<p style="color:#fff;font-size:32px;font-weight:700;letter-spacing:8px;margin:24px 0;">%s</p>
<p style="color:#666;font-size:12px;">รหัสมีอายุ 10 นาที หากคุณไม่ได้ร้องขอ กรุณาเพิกเฉยต่ออีเมลนี้</p>
</div></body></html>`, html.EscapeString(username), html.EscapeString(code))
	return m.send(toEmail, username, subject, body)
}

func (m *Mailer) send(toEmail, username, subject, htmlBody string) error {
	if !m.enabled {
		log.Printf("WARN mailer disabled, skipping email (subject: %s)", subject)
		return nil
	}

	payload := map[string]interface{}{
		"sender": map[string]string{
			"name":  m.senderName,
			"email": m.senderEmail,
		},
		"to": []map[string]string{
			{
				"email": toEmail,
				"name":  username,
			},
		},
		"subject":     subject,
		"htmlContent": htmlBody,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("mailer: marshal json payload failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("mailer: create request failed: %w", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", m.apiKey)
	req.Header.Set("content-type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mailer: brevo http request failed (subject: %s): %w", subject, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mailer: brevo api returned status %d (subject: %s): %s", resp.StatusCode, subject, string(respBody))
	}

	log.Printf("✅ Mailer: email sent successfully via Brevo (subject: %s)", subject)
	return nil
}

func buildVerifyEmailBody(username, verifyURL string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="margin:0;padding:0;background:#050505;font-family:'Helvetica Neue',Arial,sans-serif;">
  <table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background:#050505;padding:56px 16px;">
    <tr>
      <td align="center">
        <table role="presentation" width="440" cellpadding="0" cellspacing="0" style="max-width:440px;width:100%%;">

          <!-- Wordmark -->
          <tr>
            <td align="center" style="padding-bottom:32px;">
              <span style="font-size:18px;font-weight:700;letter-spacing:4px;color:#ffffff;">POPCORN</span>
              <div style="width:32px;height:2px;background:#f5432c;margin:12px auto 0 auto;"></div>
            </td>
          </tr>

          <!-- Card -->
          <tr>
            <td style="background:#0f0f0f;border:1px solid #1c1c1c;border-radius:4px;">
              <table role="presentation" width="100%%" cellpadding="0" cellspacing="0">
                <tr>
                  <td style="padding:44px 40px 4px 40px;" align="center">
                    <p style="margin:0 0 12px 0;color:#f5432c;font-size:11px;font-weight:700;letter-spacing:2.5px;
                              text-transform:uppercase;">
                      ยินดีต้อนรับ
                    </p>
                    <p style="margin:0 0 20px 0;color:#ffffff;font-size:18px;font-weight:600;line-height:1.4;">
                      สวัสดี %s
                    </p>
                    <p style="margin:0 0 32px 0;color:#8f8f8f;font-size:14px;line-height:1.8;max-width:300px;">
                      ขอบคุณที่สมัครสมาชิก POPCORN กดปุ่มด้านล่างเพื่อยืนยันอีเมล
                      และเริ่มต้นใช้งานบัญชีของคุณได้ทันที
                    </p>
                  </td>
                </tr>
                <tr>
                  <td align="center" style="padding:4px 40px 0 40px;">
                    <table role="presentation" cellpadding="0" cellspacing="0" width="100%%">
                      <tr>
                        <td style="background:#f5432c;border-radius:6px;">
                          <a href="%s"
                             style="display:block;padding:15px 44px;color:#ffffff;font-size:13px;
                                    font-weight:700;letter-spacing:1.5px;text-transform:uppercase;
                                    text-decoration:none;text-align:center;">
                            ยืนยันอีเมลของฉัน
                          </a>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>
                <tr>
                  <td align="center" style="padding:22px 40px 8px 40px;">
                    <p style="margin:0;color:#666666;font-size:12px;">
                      ลิงก์มีอายุการใช้งาน 24 ชั่วโมง
                    </p>
                  </td>
                </tr>
                <tr>
                  <td style="padding:12px 40px 40px 40px;">
                    <p style="margin:0;color:#4a4a4a;font-size:11px;line-height:1.7;text-align:center;word-break:break-all;">
                      หากปุ่มด้านบนใช้งานไม่ได้ ให้คัดลอกลิงก์นี้ไปวางในเบราว์เซอร์<br>
                      <a href="%s" style="color:#e5504f;text-decoration:underline;">%s</a>
                    </p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <!-- Footnote -->
          <tr>
            <td style="padding:24px 8px 0 8px;">
              <p style="margin:0;color:#555555;font-size:11.5px;line-height:1.7;text-align:center;">
                หากคุณไม่ได้สมัครสมาชิก POPCORN กรุณาเพิกเฉยต่ออีเมลฉบับนี้
              </p>
            </td>
          </tr>
          <tr>
            <td align="center" style="padding:28px 0 0 0;">
              <p style="margin:0;color:#3a3a3a;font-size:10.5px;letter-spacing:0.5px;">
                POPCORN — ระบบส่งอัตโนมัติ กรุณาอย่าตอบกลับอีเมลนี้
              </p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, username, verifyURL, verifyURL, verifyURL)
}
