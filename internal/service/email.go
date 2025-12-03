package service

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
	"time"

	"goboot/pkg/database"
	"goboot/pkg/logger"

	"github.com/google/uuid"
)

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

// getConfig 获取邮件配置(从数据库)
func (s *EmailService) getConfig() *EmailConfig {
	return GetConfigService().GetEmailConfig()
}

// SendMail 发送邮件
func (s *EmailService) SendMail(to, subject, body string) error {
	cfg := s.getConfig()

	if !cfg.Enabled {
		return errors.New("邮件服务未启用")
	}

	// 构建邮件头
	header := make(map[string]string)
	header["From"] = fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromAddr)
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=UTF-8"

	// 构建邮件内容
	var message strings.Builder
	for k, v := range header {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)

	if cfg.SSL {
		return s.sendMailSSL(addr, auth, cfg.FromAddr, []string{to}, []byte(message.String()), cfg.Host)
	}

	return smtp.SendMail(addr, auth, cfg.FromAddr, []string{to}, []byte(message.String()))
}

// sendMailSSL 通过 SSL 发送邮件
func (s *EmailService) sendMailSSL(addr string, auth smtp.Auth, from string, to []string, msg []byte, host string) error {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("连接SMTP服务器失败: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("创建SMTP客户端失败: %v", err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP认证失败: %v", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("设置收件人失败: %v", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("获取写入器失败: %v", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("写入邮件内容失败: %v", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("关闭写入器失败: %v", err)
	}

	return client.Quit()
}

// SendPasswordResetEmail 发送密码重置邮件
func (s *EmailService) SendPasswordResetEmail(email, username string, userID uint) error {
	cfg := s.getConfig()

	// 生成重置 token
	token := uuid.New().String()

	// 存储 token 到 Redis，设置过期时间
	ctx := context.Background()
	key := fmt.Sprintf("password_reset:%s", token)
	expire := time.Duration(cfg.ResetExpire) * time.Minute

	// 存储用户ID
	if err := database.RDB.Set(ctx, key, userID, expire).Err(); err != nil {
		return fmt.Errorf("存储重置token失败: %v", err)
	}

	// 构建重置链接
	resetLink := fmt.Sprintf("%s?token=%s", cfg.ResetURL, token)

	// 邮件内容
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2c3e50;">密码重置</h2>
        <p>您好，%s：</p>
        <p>我们收到了您的密码重置请求。请点击下面的按钮重置您的密码：</p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #3498db; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">重置密码</a>
        </p>
        <p>或者复制以下链接到浏览器：</p>
        <p style="word-break: break-all; color: #3498db;">%s</p>
        <p style="color: #e74c3c;">此链接将在 %d 分钟后失效。</p>
        <p>如果您没有请求重置密码，请忽略此邮件。</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="color: #999; font-size: 12px;">此邮件由系统自动发送，请勿回复。</p>
    </div>
</body>
</html>
`, username, resetLink, resetLink, cfg.ResetExpire)

	// 异步发送邮件
	go func() {
		if err := s.SendMail(email, "密码重置", body); err != nil {
			logger.Error("发送密码重置邮件失败", slog.String("email", email), slog.Any("error", err))
		}
	}()

	return nil
}

// VerifyResetToken 验证重置 token
func (s *EmailService) VerifyResetToken(token string) (uint, error) {
	ctx := context.Background()
	key := fmt.Sprintf("password_reset:%s", token)

	// 获取用户ID
	userIDStr, err := database.RDB.Get(ctx, key).Result()
	if err != nil {
		return 0, errors.New("重置链接无效或已过期")
	}

	var userID uint
	fmt.Sscanf(userIDStr, "%d", &userID)

	return userID, nil
}

// DeleteResetToken 删除重置 token
func (s *EmailService) DeleteResetToken(token string) error {
	ctx := context.Background()
	key := fmt.Sprintf("password_reset:%s", token)
	return database.RDB.Del(ctx, key).Err()
}

// SendNotificationEmail 发送通知邮件
func (s *EmailService) SendNotificationEmail(email, username, title, content string) error {
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2c3e50;">%s</h2>
        <p>您好，%s：</p>
        <div style="padding: 20px; background-color: #f9f9f9; border-radius: 5px; margin: 20px 0;">
            %s
        </div>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="color: #999; font-size: 12px;">此邮件由系统自动发送，请勿回复。</p>
    </div>
</body>
</html>
`, title, username, content)

	// 异步发送
	go func() {
		if err := s.SendMail(email, title, body); err != nil {
			logger.Error("发送通知邮件失败", slog.String("email", email), slog.Any("error", err))
		}
	}()

	return nil
}
