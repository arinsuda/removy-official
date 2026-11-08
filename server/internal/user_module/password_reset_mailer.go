package user_module

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/arinsuda/movie-hub/config"
	"github.com/arinsuda/movie-hub/internal/mailer"
)

type PasswordResetMailer interface {
	SendResetLink(toEmail, resetURL string) error
}

type brevoPasswordResetMailer struct {
	apiKey      string
	senderEmail string
	senderName  string
	enabled     bool
	devLog      bool
	httpClient  *http.Client
}

func NewBrevoPasswordResetMailer(cfg config.BrevoConfig) PasswordResetMailer {
	return &brevoPasswordResetMailer{
		apiKey:      cfg.APIKey,
		senderEmail: cfg.SenderEmail,
		senderName:  cfg.SenderName,
		enabled:     cfg.Enabled,
		devLog:      cfg.DevLog,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *brevoPasswordResetMailer) SendResetLink(toEmail, resetURL string) error {
	if m.devLog {
		mailer.PrintDev(os.Stdout, toEmail, "reset", "link: "+resetURL)
		return nil
	}
	if !m.enabled {
		log.Printf("WARN mailer disabled, skipping password reset email to %s", toEmail)
		return nil
	}

	subject := "รีเซ็ตรหัสผ่านบัญชีของคุณ — POPCORN"
	htmlBody := buildResetPasswordEmailBody(resetURL)

	payload := map[string]interface{}{
		"sender": map[string]string{
			"name":  m.senderName,
			"email": m.senderEmail,
		},
		"to": []map[string]string{
			{"email": toEmail},
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
		return fmt.Errorf("mailer: brevo http request send reset link to %s failed: %w", toEmail, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mailer: brevo api returned status %d for %s: %s", resp.StatusCode, toEmail, string(respBody))
	}

	log.Printf("✅ Mailer: password reset email sent successfully via Brevo to %s", toEmail)
	return nil
}

func buildResetPasswordEmailBody(resetURL string) string {
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
                      ความปลอดภัยของบัญชี
                    </p>
                    <p style="margin:0 0 32px 0;color:#8f8f8f;font-size:14px;line-height:1.8;max-width:300px;">
                      เราได้รับคำขอตั้งรหัสผ่านใหม่สำหรับบัญชีของคุณ กดปุ่มด้านล่างเพื่อดำเนินการต่อ
                    </p>
                  </td>
                </tr>
                <tr>
                  <td align="center" style="padding:4px 40px 0 40px;">
                    <table role="presentation" cellpadding="0" cellspacing="0">
                      <tr>
                        <td style="background:#f5432c;border-radius:6px;">
                          <a href="%s"
                             style="display:block;padding:15px 44px;color:#ffffff;font-size:13px;
                                    font-weight:700;letter-spacing:1.5px;text-transform:uppercase;
                                    text-decoration:none;">
                            ตั้งรหัสผ่านใหม่
                          </a>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>
                <tr>
                  <td align="center" style="padding:22px 40px 8px 40px;">
                    <p style="margin:0;color:#666666;font-size:12px;">
                      ลิงก์มีอายุการใช้งาน 1 ชั่วโมง
                    </p>
                  </td>
                </tr>
                <tr>
                  <td style="padding:12px 40px 40px 40px;">
                    <p style="margin:0;color:#4a4a4a;font-size:11px;line-height:1.7;text-align:center;word-break:break-all;">
                      หรือคัดลอกลิงก์นี้ไปวางในเบราว์เซอร์<br>
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
                หากคุณไม่ได้ขอรีเซ็ตรหัสผ่าน กรุณาเพิกเฉยต่ออีเมลฉบับนี้ รหัสผ่านเดิมของคุณจะยังคงใช้งานได้ตามปกติ
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
</html>`, resetURL, resetURL, resetURL)
}
