package main

import (
	"context"
	"encoding/json"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
	"user-management-api/internal/config"
	v1services "user-management-api/internal/services/v1"
	"user-management-api/internal/utils"
	"user-management-api/pkg/email"
	"user-management-api/pkg/logger"
	"user-management-api/pkg/rabbitmq"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

type Worker struct {
	rabbitMQ    rabbitmq.RabbitMQService
	mailService email.EmailProviderService
	cfg         *config.Config
	logger      *zerolog.Logger
}

func NewWorker(cfg *config.Config) *Worker {
	log := utils.NewLoggerWithPath("worker.log", "info")

	// connect rabbitmq
	rabbitMQ, err := rabbitmq.NewRabbitMQService(utils.GetEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"), log)
	if err != nil {
		log.Fatal().Err(err).Msg("❌ Failed to init RabbitmqService:")
	}
	// Init mail service
	mailLogger := utils.NewLoggerWithPath("email.log", "info")
	factory, err := email.NewProviderFactory(email.ProviderMailtrap)
	if err != nil {
		mailLogger.Error().Err(err).Msg("❌ Failed to create email provider factory:")
		return nil
	}
	mailService, err := email.NewMailService(cfg, mailLogger, factory)
	if err != nil {
		mailLogger.Error().Err(err).Msg("❌ Failed to create mail service:")
		return nil
	}
	return &Worker{
		rabbitMQ:    rabbitMQ,
		mailService: mailService,
		cfg:         cfg,
		logger:      log,
	}
}
func (w *Worker) Start(ctx context.Context) error {
	// tạo consumer
	const emailQueueName = "auth_email_queue"

	handler := func(body []byte) error {
		w.logger.Debug().Msgf("Received a message: %s", string(body))

		// var email email.Email
		var payloadEmail v1services.EmailPayload

		if err := json.Unmarshal(body, &payloadEmail); err != nil {
			w.logger.Error().Err(err).Msg("Failed to unmarshal email message")
			return err
		}
		// Tạo context mới với trace_id
		context := logger.WithTraceID(ctx, payloadEmail.TraceID)
		
		if err := w.mailService.SendEmail(context, payloadEmail.Email); err != nil {
			return utils.NewError("Failed to send reset password email", utils.ErrorCodeInternalServer)
		}
		w.logger.Info().Msgf("✅Email sent successfully to: %s", payloadEmail.Email.To)
		return nil
	}
	if err := w.rabbitMQ.Consume(ctx, emailQueueName, handler); err != nil {
		w.logger.Error().Err(err).Msg("❌ Failed to start consuming messages:")
		return err
	}
	w.logger.Info().Msgf("✅ worker started, waiting for messages in queue: %s", emailQueueName)
	<-ctx.Done()
	w.logger.Info().Msgf("🛑 work stopped consuming due to context cancellation: %s", emailQueueName)
	return ctx.Err()
}

func (w *Worker) Shutdown(ctx context.Context) error {
	w.logger.Info().Msg("Shut down worker ....")

	if err := w.rabbitMQ.Close(); err != nil {
		w.logger.Error().Err(err).Msg("❌ Failed to close RabbitMQ connection:")
		return err
	}

	w.logger.Info().Msg("✅RabbitMQ connection close successfully ")

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			w.logger.Error().Err(ctx.Err()).Msg("❌ Shutdown timed out:")
			return ctx.Err()
		}
	default:
	}
	w.logger.Info().Msg("✅ Worker shutdown complete")
	return nil
}

func main() {
	rootDir := utils.MustGetWorkingDir()
	logFile := filepath.Join(rootDir, "internal/logs/app.log")
	logger.InitLogger(logger.LoggerConfig{
		Level:      "info",
		Filename:   logFile,
		MaxSize:    1, // megabytes
		MaxBackups: 5,
		MaxAge:     5,    //
		Compress:   true, // disabled by default
		IsDev:      utils.GetEnv("APP_ENV", "development"),
	})

	if err := godotenv.Load(filepath.Join(rootDir, ".env")); err != nil {
		logger.Log.Warn().Msg("⚠️ No env file found")

	} else {
		logger.Log.Info().Msg("✅ Loaded successfully env in worker file: ")
	}
	// Initialize configuration
	config := config.NewConfig()
	worker := NewWorker(config)
	if worker == nil {
		logger.Log.Fatal().Msg("❌ Failed to create worker")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)
	//
	go func() {
		defer wg.Done()
		if err := worker.Start(ctx); err != nil && err != context.Canceled {
			logger.Log.Error().Err(err).Msg("❌ Worker failed to start:")
		}
	}()

	<-ctx.Done()
	logger.Log.Info().Msg("🍺 Shutdow signal receiver")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := worker.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error().Err(err).Msg("⚠️ Worker forced to shutdown:")
	}
	wg.Wait()
	logger.Log.Info().Msg("🍺 Main process terminated")
}
