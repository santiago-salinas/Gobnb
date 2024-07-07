package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/cron"

	"github.com/go-redis/redis/v8"

	"pocketbase_go/config"
	"pocketbase_go/controllers"
	logger "pocketbase_go/logger"
	repositories "pocketbase_go/repos/implementations"
	"pocketbase_go/services"
	"pocketbase_go/workers"

	"mongo-server/datasources"
	mongoRepo "mongo-server/repositories"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

var (
	monitorRequest = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "go_metrics",
		Subsystem: "prometheus",
		Name:      "Requests",
		Help:      "Counts https requests of any type",
	})

	monitorReservations = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "go_metrics",
		Subsystem: "prometheus",
		Name:      "Reservations",
		Help:      "Counts reservations",
	})

	monitorReservationPaymentSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "go_metrics",
		Subsystem: "prometheus",
		Name:      "ReservationPaymentSuccess",
		Help:      "Counts successful reservation payments",
	})

	monitorReservationPaymentFailure = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "go_metrics",
		Subsystem: "prometheus",
		Name:      "ReservationPaymentFailure",
		Help:      "Counts failed reservation payments",
	})

	opsRequested = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "go_metrics",
		Subsystem: "prometheus",
		Name:      "processed_record_count",
		Help:      "request record count",
	})
)

func metricsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		monitorRequest.Inc()
		return next(c)
	}
}

func initLogger() {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logger.Initialize("log_" + timestamp + ".log.txt")
	logger.Info("Logger initialized")
	go func() {
		logger.Debug("Got here")
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":8181", nil)
		if err != nil {
			logger.Fatal("Failed to start metrics server: ", err)
		}
	}()
}

func initMongo(mongoDatasource string) (*mongo.Client, error) {
	mongoClient, err := datasources.NewMongoDataSource(mongoDatasource)
	if err != nil {
		logger.Warn(err)
	}
	return mongoClient, err
}

func initFileServer(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/*", func(c echo.Context) error {
			path := "./public" + c.Request().URL.Path

			if fileExists(path) {
				// If it's a directory, deny access
				info, err := os.Stat(path)
				if err == nil && info.IsDir() {
					return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "403 Forbidden"})
				}

				// If it's a file, serve it
				return c.File(path)
			}

			return c.JSON(http.StatusNotAcceptable, map[string]string{"message": "404 Not Found"})
		})

		return nil
	})
}

func initRedis(redisAddress string) (*redis.Client, error) {
	var redisClient *redis.Client
	_, err := net.Dial("tcp", redisAddress)
	if err == nil {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     redisAddress,
			Password: "",
			DB:       0, // Default DB number
		})
		logger.Info("Connected to Redis")
	} else {
		logger.Warn("Redis is not available")
	}
	return redisClient, err
}

func initRabbit(rabbitAddress string) (workers.Worker, error) {
	worker, err := workers.BuildRabbitWorker(rabbitAddress)
	if err != nil {
		logger.Error(err)
	}
	return worker, err
}

func main() {
	fmt.Printf("Started")
	config.InitConfig()

	viper := config.GetConfig()
	rabbitAddress := viper.GetString("rabbit_address")
	redisAddress := viper.GetString("redis_address")
	mongoDatasource := viper.GetString("mongo_datasource")

	client_secret := viper.GetString("client_secret")
	client_id := viper.GetString("client_id")
	token_verification_url := viper.GetString("token_verification_url")

	defaultRefundPercentage := viper.GetFloat64("default_refund_percentage")
	defaultCancellationDays := viper.GetInt("default_cancellation_days")

	propertyImagesUrl := viper.GetString("property_images_url")
	propertyImagesDir := viper.GetString("property_images_dir")
	propertyImagesCompressionScale := viper.GetString("property_images_compression_scale")

	paymentURL := viper.GetString("payment_url")
	refundURL := viper.GetString("refund_url")

	initLogger()
	mongoClient, mongoErr := initMongo(mongoDatasource)
	app := pocketbase.New()
	initFileServer(app)
	redisClient, redisErr := initRedis(redisAddress)
	worker, rabbitErr := initRabbit(rabbitAddress)

	if redisErr == nil {
		defer redisClient.Close()
	}
	if mongoErr == nil {
		defer mongoClient.Disconnect(context.TODO())
	}
	if rabbitErr == nil {
		defer worker.Close()
	}

	// Repos
	reportsRepo := mongoRepo.NewReportsMongoRepo(mongoClient, "reports")
	propertyRepo := repositories.PocketPropertyRepo{Db: *app, Cache: redisClient}
	propertyRepo.SetConfigValues(propertyImagesUrl, propertyImagesDir, propertyImagesCompressionScale)
	userRepo := repositories.PocketUserRepo{Db: *app, Cache: redisClient}
	userRepo.SetConfigVariables(client_secret, client_id, token_verification_url)
	reservationsRepo := repositories.PocketReservationRepo{Db: *app}
	sensorRepo := repositories.PocketSensorRepo{Db: *app, Cache: redisClient}
	settingsRepo := repositories.PocketSettingsRepo{Db: *app}
	settingsRepo.SetConfigValues(defaultRefundPercentage, defaultCancellationDays)

	// Services
	propertyService := services.PropertyService{Repo: &propertyRepo, UserRepo: &userRepo}
	authService := services.AuthService{Repo: &userRepo}
	reservationService := services.ReservationService{ReservationRepo: &reservationsRepo, UserRepo: &userRepo, SettingsRepo: &settingsRepo, PropertiesRepo: &propertyRepo}
	reservationService.SetConfigValues(refundURL)
	sensorService := services.SensorService{Repo: &sensorRepo}
	paymentService := services.PaymentService{UsersRepo: &userRepo, PropertyRepo: &propertyRepo, ReservationRepo: &reservationsRepo}
	paymentService.SetConfigValues(paymentURL)
	reportsService := services.ReportsService{ReservationRepo: &reservationsRepo, PropertiesRepo: &propertyRepo, UsersRepo: &userRepo, ReportsRepo: reportsRepo, SensorRepo: &sensorRepo}
	notificationService := services.NewNotificationService(redisClient)

	// Controllers
	propertyController := controllers.PropertyController{Service: &propertyService, PaymentService: &paymentService, AuthService: authService}
	reservationsController := controllers.ReservationsController{ReservationsService: &reservationService, AuthService: authService, PaymentService: &paymentService}
	authController := controllers.AuthController{AuthService: authService}
	sensorController := controllers.SensorController{Service: &sensorService, AuthService: authService}
	reportsController := controllers.NewReportsController(authService, &reportsService, notificationService, worker)
	notificationsController := controllers.NewNotificationsController(notificationService, &reservationService)

	sensorController.InitSensorEndpoints(*app)
	propertyController.InitPropertyEndpoints(*app)
	reservationsController.InitReservationEndpoints(*app, monitorReservations, monitorReservationPaymentSuccess, monitorReservationPaymentFailure)
	reportsController.InitReportsEndpoints(*app)
	notificationsController.InitNotificationsEndpoints(*app)
	authController.InitAuthEndpoints(*app)

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		scheduler := cron.New()

		err := scheduler.Add("reservationDiscard", "@daily", func() {
			reservationsController.AutoCancelReservations()
		})

		if err != nil {
			logger.Error("Error scheduling job:", err)
		} else {
			scheduler.Start()
		}

		e.Router.Use(metricsMiddleware)

		return nil
	})

	// Start PocketBase server
	if err := app.Start(); err != nil {
		logger.Fatal(err)
	}
}
