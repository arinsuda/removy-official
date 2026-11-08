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

type Mailer interface {
	SendOTP(toEmail, otp string) error
}

type brevoMailer struct {
	apiKey      string
	senderEmail string
	senderName  string
	enabled     bool
	devLog      bool
	httpClient  *http.Client
}

func NewBrevoMailer(cfg config.BrevoConfig) Mailer {
	return &brevoMailer{
		apiKey:      cfg.APIKey,
		senderEmail: cfg.SenderEmail,
		senderName:  cfg.SenderName,
		enabled:     cfg.Enabled,
		devLog:      cfg.DevLog,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *brevoMailer) SendOTP(toEmail, otp string) error {
	if m.devLog {
		mailer.PrintDev(os.Stdout, toEmail, "email_change_otp", "code: "+otp)
		return nil
	}
	if !m.enabled {
		log.Printf("WARN mailer disabled, skipping OTP email to %s", toEmail)
		return nil
	}

	subject := "รหัสยืนยันการเปลี่ยนอีเมล — POPCORN"
	htmlBody := buildOTPEmailBody(otp)

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
		return fmt.Errorf("mailer: brevo http request send otp to %s failed: %w", toEmail, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mailer: brevo api returned status %d for %s: %s", resp.StatusCode, toEmail, string(respBody))
	}

	log.Printf("✅ Mailer: OTP email sent successfully via Brevo to %s", toEmail)
	return nil
}

func buildOTPEmailBody(otp string) string {
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
                      คำขอเปลี่ยนอีเมล
                    </p>
                    <p style="margin:0 0 28px 0;color:#8f8f8f;font-size:14px;line-height:1.8;max-width:300px;">
                      กรอกรหัสยืนยันด้านล่างเพื่อดำเนินการเปลี่ยนอีเมลของบัญชีคุณ
                    </p>
                  </td>
                </tr>

                <!-- OTP code: copy-friendly, ระยะห่างทำด้วย letter-spacing ล้วน text สะอาด -->
                <tr>
                  <td align="center" style="padding:0 40px;">
                    <table role="presentation" cellpadding="0" cellspacing="0"
                           style="background:#050505;border:1px solid #f5432c;border-radius:6px;">
                      <tr>
                        <td style="padding:20px 32px;">
                          <span style="font-size:30px;font-weight:700;letter-spacing:10px;color:#ffffff;
                                       font-family:'Courier New',monospace;user-select:all;">%s</span>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>

                <tr>
                  <td align="center" style="padding:14px 40px 4px 40px;">
                    <p style="margin:0;color:#4a4a4a;font-size:11px;line-height:1.6;">
                      แตะ/คลิกที่รหัสด้านบนค้างไว้เพื่อเลือกทั้งหมด แล้วคัดลอกได้ทันที
                    </p>
                  </td>
                </tr>
                <tr>
                  <td align="center" style="padding:16px 40px 40px 40px;">
                    <p style="margin:0;color:#666666;font-size:12px;">
                      รหัสมีอายุการใช้งาน 15 นาที
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
                หากคุณไม่ได้เป็นผู้ทำรายการนี้ กรุณาเพิกเฉยต่ออีเมลฉบับนี้
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
</html>`, otp)
}
