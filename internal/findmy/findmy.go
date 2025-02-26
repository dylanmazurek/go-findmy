package findmy

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dylanmazurek/go-findmy/internal/publisher"
	pubModels "github.com/dylanmazurek/go-findmy/internal/publisher/models"
	"github.com/dylanmazurek/go-findmy/pkg/notifier"
	"github.com/dylanmazurek/go-findmy/pkg/nova"
	"github.com/dylanmazurek/go-findmy/pkg/nova/models/protos/bindings"
	"github.com/dylanmazurek/go-findmy/pkg/shared/vault"
	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"
)

type FindMy struct {
	novaClient      *nova.Client
	notifyClient    *notifier.Client
	publisherClient *publisher.Client

	vaultClient       *vault.Client
	internalScheduler gocron.Scheduler
}

func NewFindMy() (*FindMy, error) {
	ctx := context.Background()

	var newFindMy FindMy
	err := newFindMy.initClients(ctx)
	if err != nil {
		return nil, err
	}

	timezone, err := time.LoadLocation("Australia/Melbourne")
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

	job := gocron.CronJob("*/20 * * * *", false)
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

func (f *FindMy) GetDevices(ctx context.Context) []pubModels.Device {
	log := log.Ctx(ctx)

	devices, err := f.novaClient.GetDevices(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get devices")
	}

	var pubDevices []pubModels.Device
	for _, device := range devices.DeviceMetadata {
		if device.GetIdentifierInformation().GetType() != bindings.IdentifierInformationType_IDENTIFIER_ANDROID {
			deviceName := device.GetUserDefinedDeviceName()
			model := device.GetInformation().GetDeviceRegistration().GetModel()
			manufacturer := device.GetInformation().GetDeviceRegistration().GetManufacturer()

			canonicId := device.GetIdentifierInformation().GetCanonicIds().GetCanonicId()[0].GetId()
			canonicIdSplit := strings.Split(canonicId, "-")
			serial := canonicIdSplit[len(canonicIdSplit)-1]

			newPubDevice := pubModels.NewDevice(deviceName, serial, model, manufacturer)
			pubDevices = append(pubDevices, newPubDevice)
		}
	}

	return pubDevices
}
