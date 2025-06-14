package findmy

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dylanmazurek/go-findmy/internal/findmy/constants"
	"github.com/dylanmazurek/go-findmy/internal/publisher"
	"github.com/dylanmazurek/go-findmy/pkg/notifier"
	"github.com/dylanmazurek/go-findmy/pkg/nova"
	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"
)

type Service struct {
	novaClient      *nova.Client
	notifierClient  *notifier.Client
	publisherClient *publisher.Client

	internalScheduler gocron.Scheduler
}

func NewFindMy(ctx context.Context) (*Service, error) {
	log := log.Ctx(ctx).With().Str("service", constants.SERVICE_NAME).Logger()

	log.Debug().Msg("creating new find-my service")

	var newFindMyService Service

	err := newFindMyService.initClients(ctx)
	if err != nil {
		return nil, err
	}

	timezoneEnv, hasTimezoneEnv := os.LookupEnv("TIMEZONE")
	if !hasTimezoneEnv {
		timezoneEnv = constants.DEFAULT_TIMEZONE
	}

	timezone, err := time.LoadLocation(timezoneEnv)
	if err != nil {
		return nil, err
	}

	opts := []gocron.SchedulerOption{
		gocron.WithLocation(timezone),
		gocron.WithLimitConcurrentJobs(1, gocron.LimitModeWait),
	}

	scheduler, err := gocron.NewScheduler(opts...)
	if err != nil {
		return nil, err
	}

	newFindMyService.internalScheduler = scheduler

	return &newFindMyService, nil
}

func (s *Service) AddJobs(ctx context.Context) error {
	log := log.Ctx(ctx)

	cronSchedule, hasCronSchedule := os.LookupEnv("CRON_SCHEDULE")
	if !hasCronSchedule {
		cronSchedule = constants.DEFAULT_CRON_SCHEDULE
	}

	job := gocron.CronJob(cronSchedule, false)
	task := gocron.NewTask(func(ctx context.Context) {
		err := s.novaClient.RefreshDevices(ctx)
		if err != nil {
			log.Error().Err(err).Msg("failed to get devices")
		}
	}, ctx)

	jobOpts := []gocron.JobOption{
		gocron.WithStartAt(gocron.WithStartImmediately()),
	}

	newJob, err := s.internalScheduler.NewJob(job, task, jobOpts...)
	if err != nil {
		return err
	}

	log.Info().
		Str("job_id", newJob.ID().String()).
		Msg("job added")

	return nil
}

func (s *Service) Start(ctx context.Context) error {
	log := log.Ctx(ctx).With().Str("service", constants.SERVICE_NAME).Logger()

	devices, err := s.GetDevices(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get devices")
	}

	s.publisherClient.InitalizeDevices(ctx, devices)

	log.Trace().Msg("starting find-my service")

	err = s.notifierClient.StartListening(ctx)
	if err != nil {
		return err
	}

	s.AddJobs(ctx)

	log.Debug().Msg("starting scheduler")

	s.internalScheduler.Start()

	log.Debug().Msg("scheduler started")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Msg("find-my service running")

	<-sigs

	log.Info().Msg("received terminate signal, stopping listener")

	return err
}
