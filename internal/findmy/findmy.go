package findmy

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dylanmazurek/go-findmy/internal/publisher"
	"github.com/dylanmazurek/go-findmy/pkg/notifier"
	"github.com/dylanmazurek/go-findmy/pkg/nova"
	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"
)

type FindMy struct {
	novaClient      *nova.Client
	notifyClient    *notifier.Client
	publisherClient *publisher.Client

	internalScheduler gocron.Scheduler
}

func NewFindMy() (*FindMy, error) {
	ctx := context.Background()

	var newFindMy FindMy
	err := newFindMy.initClients(ctx)
	if err != nil {
		return nil, err
	}

	timezoneEnv, hasTimezoneEnv := os.LookupEnv("TIMEZONE")
	if !hasTimezoneEnv {
		timezoneEnv = "Australia/Melbourne"
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

	newFindMy.internalScheduler = scheduler

	return &newFindMy, nil
}

func (f *FindMy) AddJobs(ctx context.Context) error {
	log := log.Ctx(ctx)

	cronSchedule, hasCronSchedule := os.LookupEnv("CRON_SCHEDULE")
	if !hasCronSchedule {
		cronSchedule = "*/20 * * * *"
	}

	job := gocron.CronJob(cronSchedule, false)
	task := gocron.NewTask(func(ctx context.Context) {
		log.Info().Msg("refreshing devices")
		novaClient := f.novaClient
		err := novaClient.RefreshAllDevices(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get devices")
		}
	}, ctx)

	jobOpts := []gocron.JobOption{
		gocron.WithStartAt(gocron.WithStartImmediately()),
	}

	newJob, err := f.internalScheduler.NewJob(job, task, jobOpts...)
	if err != nil {
		return err
	}

	log.Info().Str("job_id", string(newJob.ID().String())).Msg("job added")

	return nil
}

func (f *FindMy) Start(ctx context.Context) error {
	log := log.Ctx(ctx)
	log.Info().Msg("starting to listen for notifications")
	err := f.notifyClient.StartListening(ctx)
	if err != nil {
		return err
	}

	f.AddJobs(ctx)

	log.Info().Msg("starting scheduler")
	f.internalScheduler.Start()

	log.Info().Msg("listening for messages")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Msg("waiting for signals")

	<-sigs

	log.Info().Msg("received signal, stopping listener")

	return err
}
