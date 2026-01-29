package backup

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
)

// Scheduler runs periodic backup and cleanup jobs.
type Scheduler struct {
	cron    *cron.Cron
	service *Service
	logger  zerolog.Logger
}

// NewScheduler creates a new backup scheduler.
func NewScheduler(service *Service, logger zerolog.Logger) *Scheduler {
	return &Scheduler{
		cron:    cron.New(),
		service: service,
		logger:  logger,
	}
}

// Start adds the default backup and cleanup schedules and starts the cron runner.
// backupCron defaults to "0 */6 * * *" (every 6 hours).
// cleanupCron defaults to "0 3 * * *" (daily at 3 AM).
func (s *Scheduler) Start(backupCron, cleanupCron string) error {
	if backupCron == "" {
		backupCron = "0 */6 * * *"
	}
	if cleanupCron == "" {
		cleanupCron = "0 3 * * *"
	}

	if _, err := s.cron.AddFunc(backupCron, func() {
		s.logger.Info().Msg("scheduled backup starting")
		if err := s.service.BackupAllTenants(context.Background()); err != nil {
			s.logger.Error().Err(err).Msg("scheduled backup failed")
		}
	}); err != nil {
		return err
	}

	if _, err := s.cron.AddFunc(cleanupCron, func() {
		s.logger.Info().Msg("scheduled cleanup starting")
		if err := s.service.CleanupExpired(context.Background()); err != nil {
			s.logger.Error().Err(err).Msg("scheduled cleanup failed")
		}
	}); err != nil {
		return err
	}

	s.cron.Start()
	s.logger.Info().Str("backup_cron", backupCron).Str("cleanup_cron", cleanupCron).Msg("backup scheduler started")
	return nil
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() context.Context {
	return s.cron.Stop()
}
