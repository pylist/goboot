package service

import (
	"context"
	"log/slog"
	"sync"

	"goboot/pkg/logger"

	"github.com/robfig/cron/v3"
)

// CronService 定时任务服务
type CronService struct {
	cron *cron.Cron
	jobs map[string]cron.EntryID
	mu   sync.RWMutex
}

// JobFunc 任务执行函数类型
type JobFunc func()

// cronService 全局单例
var cronService *CronService
var cronOnce sync.Once

// GetCronService 获取定时任务服务单例
func GetCronService() *CronService {
	cronOnce.Do(func() {
		cronService = &CronService{
			cron: cron.New(cron.WithSeconds(), cron.WithLogger(&cronLogger{})),
			jobs: make(map[string]cron.EntryID),
		}
	})
	return cronService
}

// cronLogger 适配器，实现 cron.Logger 接口
type cronLogger struct{}

func (l *cronLogger) Info(msg string, keysAndValues ...interface{}) {
	logger.Debug(msg, convertToSlogAttrs(keysAndValues)...)
}

func (l *cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	attrs := append([]any{slog.Any("error", err)}, convertToSlogAttrs(keysAndValues)...)
	logger.Error(msg, attrs...)
}

// convertToSlogAttrs 将 key-value 对转换为 slog.Attr
func convertToSlogAttrs(keysAndValues []interface{}) []any {
	attrs := make([]any, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}
		attrs = append(attrs, slog.Any(key, keysAndValues[i+1]))
	}
	return attrs
}

// Start 启动定时任务调度器
func (s *CronService) Start() {
	s.cron.Start()
	logger.Info("Cron scheduler started")
}

// Stop 停止定时任务调度器（等待正在运行的任务完成）
func (s *CronService) Stop() context.Context {
	ctx := s.cron.Stop()
	logger.Info("Cron scheduler stopped")
	return ctx
}

// AddJob 添加定时任务
// name: 任务名称（唯一标识）
// spec: cron 表达式（支持秒级，格式：秒 分 时 日 月 周）
// job: 任务执行函数
func (s *CronService) AddJob(name, spec string, job JobFunc) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果任务已存在，先移除
	if entryID, exists := s.jobs[name]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, name)
	}

	// 包装任务函数，添加日志和 panic 恢复
	wrappedJob := func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Cron job panic",
					slog.String("job", name),
					slog.Any("panic", r),
				)
			}
		}()

		logger.Debug("Cron job executing", slog.String("job", name))
		job()
		logger.Debug("Cron job completed", slog.String("job", name))
	}

	entryID, err := s.cron.AddFunc(spec, wrappedJob)
	if err != nil {
		logger.Error("Failed to add cron job",
			slog.String("job", name),
			slog.String("spec", spec),
			slog.Any("error", err),
		)
		return err
	}

	s.jobs[name] = entryID
	logger.Info("Cron job added",
		slog.String("job", name),
		slog.String("spec", spec),
	)
	return nil
}

// RemoveJob 移除定时任务
func (s *CronService) RemoveJob(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	entryID, exists := s.jobs[name]
	if !exists {
		return false
	}

	s.cron.Remove(entryID)
	delete(s.jobs, name)
	logger.Info("Cron job removed", slog.String("job", name))
	return true
}

// GetJobs 获取所有任务名称
func (s *CronService) GetJobs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.jobs))
	for name := range s.jobs {
		names = append(names, name)
	}
	return names
}

// GetEntries 获取所有任务条目信息
func (s *CronService) GetEntries() []cron.Entry {
	return s.cron.Entries()
}
