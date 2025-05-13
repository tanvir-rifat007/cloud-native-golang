// Package main is the entry point to the server. It reads configuration, sets up logging and error handling,
// handles signals from the OS, and starts and stops the server.
package main

import (
	"canvas/internal/data"
	"canvas/internal/data/mailer"
	"canvas/messaging"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/smithy-go/logging"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"go.uber.org/zap"
)

var release string


type cfg struct{
	Port string
	Env  string
	db   struct{
		DSN string
	}
}

type application struct{
	logger *zap.Logger
	config cfg
	models data.Model
	queue  *messaging.Queue

	blobstore *data.BlobStore

	registry *prometheus.Registry
	requests         *prometheus.CounterVec
	requestDurations *prometheus.HistogramVec
	jobCount         *prometheus.CounterVec
	jobDurations     *prometheus.CounterVec
	runnerReceives    *prometheus.CounterVec


}

func main(){

  env := os.Getenv("ENV")
	fmt.Print("env nai:",env)
	if env == "" {
		env = "development"
	}

	// Load .env only in development
	if env == "development" {
		if err := godotenv.Load(); err != nil {
			fmt.Println("Warning: .env file not found")
		}
	}

	


	var dsn string

	if (env == "development") {
dsn = fmt.Sprintf(
	"postgres://%s:%s@%s/%s?sslmode=disable",
	os.Getenv("DB_USER"),
	os.Getenv("DB_PASSWORD"),
	os.Getenv("DB_HOST"),
	os.Getenv("DB_NAME"),
)
	}

	if (env == "production") {
		dsn = fmt.Sprintf(
	"postgres://%s:%s@%s/%s?sslmode=require",
	os.Getenv("DB_USER"),
	os.Getenv("DB_PASSWORD"),
	os.Getenv("DB_HOST"),
	os.Getenv("DB_NAME"),
)
	}

	

fmt.Println(dsn)

	var cfg cfg

	flag.StringVar(&cfg.Port, "port",os.Getenv("PORT"), "API server port")

	flag.StringVar(&cfg.Env, "env", os.Getenv("ENV"), "Environment (development|production)")

	flag.StringVar(&cfg.db.DSN, "dsn", dsn, "PostgreSQL DSN")

	flag.Parse()

	// create a zap logger
	log,err:= createLogger(cfg.Env)

	if err != nil {
		fmt.Println("Error creating logger:", err)
		return
	}

	log = log.With(zap.String("release", release))

	defer func(){
		_ = log.Sync()

	}()


		// open a connection to the database
	db, err := openDB(cfg)

	if err != nil {
		log.Error("Error opening database connection", zap.Error(err))
		os.Exit(1)
	}

	defer db.Close()

	log.Info("Database connection established", zap.String("dsn", cfg.db.DSN))

	fmt.Println("AWS_ACCESS_KEY_ID:", os.Getenv("AWS_ACCESS_KEY_ID"))
fmt.Println("AWS_SECRET_ACCESS_KEY:", os.Getenv("AWS_SECRET_ACCESS_KEY"))


awsConfig, err := config.LoadDefaultConfig(context.Background(),
config.WithLogger(createAWSLogAdapter(log)),
config.WithEndpointResolver(createAWSEndpointResolver()),
)
if err != nil {
	log.Error("Error loading AWS config", zap.Error(err))
	return

}


	if err != nil {
		log.Info("Error creating AWS config", zap.Error(err))
		return 
	}




	registry := prometheus.NewRegistry()

	// http metrics
requests := promauto.With(registry).NewCounterVec(prometheus.CounterOpts{
	Name: "app_http_requests_total",
	Help: "The total number of HTTP requests.",
}, []string{"method", "path", "code"})

requestDurations := promauto.With(registry).NewHistogramVec(prometheus.HistogramOpts{
	Name:    "app_http_request_duration_seconds",
	Help:    "HTTP request durations.",
	Buckets: []float64{.005, .01, .05, .1, .5, 1},
}, []string{"code"})

// databse metrics
dbCollector := collectors.NewDBStatsCollector(db, "db")
err = registry.Register(dbCollector)
if err != nil {
	log.Warn("DB stats collector already registered?", zap.Error(err))
}



// for job metrics that pool the message queue
jobCount := promauto.With(registry).NewCounterVec(prometheus.CounterOpts{
	Name: "app_jobs_total",
	Help: "The total number of jobs processed",
}, []string{"name", "success"})

jobDurations := promauto.With(registry).NewCounterVec(prometheus.CounterOpts{
	Name: "app_job_duration_seconds_total",
	Help: "The total duration of jobs in seconds",
}, []string{"name", "success"})

runnerReceives := promauto.With(registry).NewCounterVec(prometheus.CounterOpts{
	Name: "app_job_runner_receives_total",
	Help: "The number of times the job runner received a message",
}, []string{"success"})



	

	// create an application struct

	app := &application{
		logger: log,
		config: cfg,
		models: data.NewModel(db),
		queue:    createQueue(log, awsConfig),
		blobstore: createBlobStore(log, awsConfig),
		registry: registry,
		requests:         requests,
		requestDurations: requestDurations,
		jobCount: 			 jobCount,
		jobDurations:     jobDurations,
		runnerReceives:    runnerReceives,

	}

	// Start background worker
ctx := context.Background()
app.startEmailWorker(ctx)






	err = app.serve(":" + cfg.Port)

	if err != nil {
		app.logger.Error("Error starting server", zap.Error(err))
		os.Exit(1)
	}

}

func createLogger(env string) (*zap.Logger,error){
	switch env{
		case "development":
			return zap.NewDevelopment()
		case "production":
			return zap.NewProduction()
		
		default:
			return zap.NewNop(),nil
	}
}


func openDB(cfg cfg)(*sql.DB,error){
	db,err:= sql.Open("postgres",cfg.db.DSN)

	if err!=nil{
		return nil,err
	}

	ctx,cancel:= context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)

	if err!=nil{
		return nil,err
	}



	return db,nil
}


func createAWSLogAdapter(log *zap.Logger) logging.LoggerFunc {
	return func(classification logging.Classification, format string, v ...interface{}) {
		switch classification {
		case logging.Debug:
			log.Sugar().Debugf(format, v...)
		case logging.Warn:
			log.Sugar().Warnf(format, v...)
		}
	}
}

// createAWSEndpointResolver used for local development endpoints.
// See https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/endpoints/
// func createAWSEndpointResolver() aws.EndpointResolverFunc {
//     sqsEndpointURL := os.Getenv("SQS_ENDPOINT_URL")
//     s3EndpointURL := os.Getenv("S3_ENDPOINT_URL")
//     awsRegion := os.Getenv("AWS_REGION")

//     if awsRegion == "" {
//         awsRegion = "eu-north-1" // Default region if not set
//     }

//     return func(service, region string) (aws.Endpoint, error) {
//         switch service {
//         case sqs.ServiceID:
//             if sqsEndpointURL != "" {
//                 return aws.Endpoint{
//                     URL:           sqsEndpointURL,
//                     SigningRegion: awsRegion,
//                 }, nil
//             }
//             // Fallback to AWS SQS regional endpoint
//             return aws.Endpoint{
//                 URL:           fmt.Sprintf("https://sqs.%s.amazonaws.com", awsRegion),
//                 SigningRegion: awsRegion,
//             }, nil

//         case s3.ServiceID:
//             if s3EndpointURL != "" {
//                 // MinIO endpoint
//                 return aws.Endpoint{
//                     URL:           s3EndpointURL,
//                     SigningRegion: awsRegion,
//                     HostnameImmutable: true, // Important for MinIO
//                 }, nil
//             }
//             // AWS S3 regional endpoint
						

//             return aws.Endpoint{
//                 URL:           fmt.Sprintf("https://%s.s3.%s.amazonaws.com", os.Getenv("BUCKET_NAME"), awsRegion),
//                 SigningRegion: awsRegion,
//                 HostnameImmutable: false,
//             }, nil

//         default:
//             return aws.Endpoint{}, &aws.EndpointNotFoundError{}
//         }
//     }
// }


// func createAWSEndpointResolver() aws.EndpointResolverWithOptionsFunc {
// 	sqsEndpointURL := os.Getenv("SQS_ENDPOINT_URL")
// 	s3EndpointURL := os.Getenv("S3_ENDPOINT_URL")
// 	awsRegion := os.Getenv("AWS_REGION")

// 	if awsRegion == "" {
// 		awsRegion = "eu-north-1"
// 	}

// 	return func(service, region string, options ...interface{}) (aws.Endpoint, error) {
// 		switch service {
// 		case sqs.ServiceID:
// 			if sqsEndpointURL != "" {
// 				return aws.Endpoint{
// 					URL:               sqsEndpointURL,
// 					SigningRegion:     awsRegion,
// 					HostnameImmutable: true,
// 				}, nil
// 			}
// 			return aws.Endpoint{
// 				URL:           fmt.Sprintf("https://sqs.%s.amazonaws.com", awsRegion),
// 				SigningRegion: awsRegion,
// 			}, nil

// 		case s3.ServiceID:
// 			if s3EndpointURL != "" {
// 				return aws.Endpoint{
// 					URL:               s3EndpointURL,
// 					SigningRegion:     awsRegion,
// 					HostnameImmutable: true,
// 				}, nil
// 			}

// 			bucketName := os.Getenv("BUCKET_NAME")
// 			if bucketName == "" {
// 				return aws.Endpoint{}, fmt.Errorf("BUCKET_NAME not set")
// 			}
// 			return aws.Endpoint{
// 				URL:               fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucketName, awsRegion),
// 				SigningRegion:     awsRegion,
// 				HostnameImmutable: false,
// 			}, nil

// 		default:
// 			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
// 		}
// 	}
// }

func createAWSEndpointResolver() aws.EndpointResolverFunc {
	sqsEndpointURL := os.Getenv("SQS_ENDPOINT_URL")
     s3EndpointURL := os.Getenv("S3_ENDPOINT_URL")

	return func(service, region string) (aws.Endpoint, error) {
		if sqsEndpointURL != "" && service == sqs.ServiceID {
			return aws.Endpoint{
				URL: sqsEndpointURL,
			}, nil
		}

		if s3EndpointURL != "" && service == s3.ServiceID {
			return aws.Endpoint{
				URL: s3EndpointURL,
			}, nil

		}

		// Fallback to default endpoint
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	}
}


// …

func createQueue(log *zap.Logger, awsConfig aws.Config) *messaging.Queue {
	return messaging.NewQueue(messaging.NewQueueOptions{
		Config:   awsConfig,
		Log:      log,
		Name:     os.Getenv("QUEUE_NAME"), // jobs
		
		
		WaitTime: 20 * time.Second,
	})
}

// not using the job runner metrics for now

// func (app *application) startEmailWorker(ctx context.Context) {
// 	app.logger.Info("Starting email worker")

// 	// mailer := mailer.NewMailer(
// 	// 	os.Getenv("SMTP_HOST"),
// 	// 	os.Getenv("SMTP_PORT"),
// 	// 	os.Getenv("SMTP_USERNAME"),
// 	// 	os.Getenv("SMTP_PASSWORD"),
// 	// 	"Canvas <canvas@tanvirrifat.me>",
// 	// )

// 	go func() {
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				app.logger.Info("Shutting down email worker")
// 				return
// 			default:
// 				// Poll the queue
// 				msg, receiptID, err := app.queue.Receive(ctx)
// 				if err != nil {
// 					app.logger.Error("Queue receive failed", zap.Error(err))
// 					time.Sleep(5 * time.Second) // backoff
// 					continue
// 				}
// 				if msg == nil {
// 					continue // nothing to process
// 				}

// 				jobType := (*msg)["job"]
// 				switch jobType {
// 				case "confirmation_email":
// 					email := (*msg)["email"]
// 					token := (*msg)["token"]

// 					app.logger.Info("Processing confirmation email", zap.String("email", email))

// 					data := map[string]any{"token": token}
// 					err = mailer.Send([]string{email}, "user_welcome.tmpl.html", "./templates/user_welcome.tmpl.html", data)
// 					if err != nil {
// 						app.logger.Error("Failed to send email", zap.Error(err))
// 						continue // don't delete message
// 					}

// 					err = app.queue.Delete(ctx, receiptID)
// 					if err != nil {
// 						app.logger.Error("Failed to delete message", zap.Error(err))
// 					}

// 				default:
// 					app.logger.Warn("Unknown job type", zap.String("job", jobType))
// 				}
// 			}
// 		}
// 	}()
// }



// using the job runner metrics to pool the message from queue:

func (app *application) startEmailWorker(ctx context.Context) {
	app.logger.Info("Starting email worker")

	
	go func() {
		for {
			select {
			case <-ctx.Done():
				app.logger.Info("Shutting down email worker")
				return
			default:
				start := time.Now()

				msg, receiptID, err := app.queue.Receive(ctx)
				if err != nil {
					app.runnerReceives.WithLabelValues("false").Inc()
					app.logger.Error("Queue receive failed", zap.Error(err))
					time.Sleep(5 * time.Second) // backoff
					continue
				}

				if msg == nil {
					app.runnerReceives.WithLabelValues("true").Inc()
					continue
				}

				app.runnerReceives.WithLabelValues("true").Inc()

				jobType := (*msg)["job"]
				success := "true"

				switch jobType {
				case "confirmation_email":
					email := (*msg)["email"]
					token := (*msg)["token"]

					app.logger.Info("Processing confirmation email", zap.String("email", email))

					data := map[string]any{"token": token}
					err := mailer.Send([]string{email}, "user_confirmation.tmpl.html", "./templates/user_confirmation.tmpl.html", data)
					if err != nil {
						app.logger.Error("Failed to send email", zap.Error(err))
						success = "false"
					}

				case "welcome_email":
	         email := (*msg)["email"]

					 // sending the user an image as a welcome email

					 giftUrl,err:=app.blobstore.CreateAndSaveNewsletterGift(ctx,email)
					 if err!=nil{
						app.logger.Error("Failed to create gift", zap.Error(err))
						success = "false"
						return

					 }
					 app.logger.Info("Gift URL:", zap.String("gift_url", giftUrl))




	         app.logger.Info("Processing welcome email", zap.String("email", email))

					 data:= map[string]any{
						"gift_url": giftUrl,
						"email":    email,
						
					 }

	        err = mailer.Send([]string{email}, "user_welcome.tmpl.html", "./templates/user_welcome.tmpl.html", data)
         	if err != nil {
	       	app.logger.Error("Failed to send welcome email", zap.Error(err))
		       success = "false"
	      }

				case "newsletter_email":
	email := (*msg)["email"]
	newsletterIDStr := (*msg)["id"]

	app.logger.Info("Processing newsletter email", zap.String("email", email), zap.String("newsletter_id", newsletterIDStr))

	// Parse newsletter ID from string to int
	newsletterID, err := strconv.Atoi(newsletterIDStr)
	if err != nil {
		app.logger.Error("Invalid newsletter ID", zap.String("id", newsletterIDStr), zap.Error(err))
		success = "false"
		break
	}

	newsletter, err := app.models.Newsletter.GetNewsletter(newsletterID)
	if err != nil {
		app.logger.Error("Failed to fetch newsletter", zap.Int("id", newsletterID), zap.Error(err))
		success = "false"
		break
	}

	data := map[string]any{
		"title": newsletter.Title,
		"body":  newsletter.Body,
	}

	err = mailer.Send([]string{email}, "newsletter.tmpl.html", "./templates/newsletter.tmpl.html", data)
	if err != nil {
		app.logger.Error("Failed to send newsletter email", zap.String("email", email), zap.Error(err))
		success = "false"
	}


				default:
					app.logger.Warn("Unknown job type", zap.String("job", jobType))
					success = "false"
				}

				duration := time.Since(start).Seconds()
				app.jobCount.WithLabelValues(jobType, success).Inc()
				app.jobDurations.WithLabelValues(jobType, success).Add(duration)

				if success == "true" {
					err = app.queue.Delete(ctx, receiptID)
					if err != nil {
						app.logger.Error("Failed to delete message", zap.Error(err))
					}
				}
			}
		}
	}()
}


func createBlobStore(log *zap.Logger, awsConfig aws.Config) *data.BlobStore {
	return data.NewBlobStore(data.NewBlobStoreOptions{
		Bucket: os.Getenv("BUCKET_NAME"),
		

		Config:    awsConfig,
		Log:       log,
		PathStyle: os.Getenv("BLOB_STORE_PATH_STYLE") == "true",
	})
}





