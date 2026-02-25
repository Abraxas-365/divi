package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Abraxas-365/divi/pkg/config"
	"github.com/Abraxas-365/divi/pkg/diveinspect/diveinspectcontainer"
	"github.com/Abraxas-365/divi/pkg/fsx"
	"github.com/Abraxas-365/divi/pkg/fsx/fsxlocal"
	"github.com/Abraxas-365/divi/pkg/fsx/fsxs3"
	"github.com/Abraxas-365/divi/pkg/iam/iamcontainer"
	"github.com/Abraxas-365/divi/pkg/logx"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	// manifesto:container-imports
)

// Container holds shared infrastructure and composed module containers.
type Container struct {
	Config *config.Config

	// Infrastructure
	DB         *sqlx.DB
	Redis      *redis.Client
	FileSystem fsx.FileSystem
	S3Client   *s3.Client

	// Bounded-context containers
	IAM         *iamcontainer.Container
	DiveInspect *diveinspectcontainer.Container
	// manifesto:container-fields
}

func NewContainer(cfg *config.Config) *Container {
	logx.Info("🔧 Initializing application container...")

	c := &Container{Config: cfg}

	c.initInfrastructure()
	c.initModules()

	logx.Info("✅ Application container initialized")
	return c
}

// ---------------------------------------------------------------------------
// Infrastructure
// ---------------------------------------------------------------------------

func (c *Container) initInfrastructure() {
	logx.Info("🏗️ Initializing infrastructure...")

	// Database
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Config.Database.Host,
		c.Config.Database.Port,
		c.Config.Database.User,
		c.Config.Database.Password,
		c.Config.Database.Name,
		c.Config.Database.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		logx.Fatalf("Failed to connect to database: %v", err)
	}
	db.SetMaxOpenConns(c.Config.Database.MaxOpenConns)
	db.SetMaxIdleConns(c.Config.Database.MaxIdleConns)
	db.SetConnMaxLifetime(c.Config.Database.ConnMaxLifetime)
	c.DB = db
	logx.Info("  ✅ Database connected")

	// Redis
	c.Redis = redis.NewClient(&redis.Options{
		Addr:     c.Config.Redis.Address(),
		Password: c.Config.Redis.Password,
		DB:       c.Config.Redis.DB,
	})
	if _, err := c.Redis.Ping(context.Background()).Result(); err != nil {
		logx.Fatalf("Failed to connect to Redis: %v (Redis is required)", err)
	}
	logx.Info("  ✅ Redis connected")

	// File storage
	c.initFileStorage()

	logx.Info("✅ Infrastructure initialized")
}

func (c *Container) initFileStorage() {
	storageMode := getEnv("STORAGE_MODE", "local")

	switch storageMode {
	case "s3":
		awsRegion := getEnv("AWS_REGION", c.Config.Email.AWSRegion)
		awsBucket := getEnv("AWS_BUCKET", "divi-uploads")

		cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(awsRegion))
		if err != nil {
			logx.Fatalf("Unable to load AWS SDK config: %v", err)
		}
		c.S3Client = s3.NewFromConfig(cfg)
		c.FileSystem = fsxs3.NewS3FileSystem(c.S3Client, awsBucket, "")
		logx.Infof("  ✅ S3 file system configured (bucket: %s, region: %s)", awsBucket, awsRegion)

	case "local":
		uploadDir := getEnv("UPLOAD_DIR", "./uploads")
		localFS, err := fsxlocal.NewLocalFileSystem(uploadDir)
		if err != nil {
			logx.Fatalf("Failed to initialize local file system: %v", err)
		}
		c.FileSystem = localFS
		logx.Infof("  ✅ Local file system configured (path: %s)", localFS.GetBasePath())

	default:
		logx.Fatalf("Unknown STORAGE_MODE: %s (use 'local' or 's3')", storageMode)
	}
}

// ---------------------------------------------------------------------------
// Module composition
// ---------------------------------------------------------------------------

func (c *Container) initModules() {
	logx.Info("📦 Initializing modules...")
	c.IAM = iamcontainer.New(iamcontainer.Deps{
		DB:          c.DB,
		Redis:       c.Redis,
		Cfg:         c.Config,
		OTPNotifier: NewConsoleNotifier(),
	})

	c.DiveInspect = diveinspectcontainer.New(diveinspectcontainer.Deps{
		DB:         c.DB,
		FileSystem: c.FileSystem,
	})

	// manifesto:module-init
}

// ---------------------------------------------------------------------------
// Lifecycle
// ---------------------------------------------------------------------------

func (c *Container) StartBackgroundServices(ctx context.Context) {
	logx.Info("🔄 Starting background services...")
	c.IAM.StartBackgroundServices(ctx)
	// manifesto:background-start
}

func (c *Container) Cleanup() {
	logx.Info("🧹 Cleaning up resources...")

	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			logx.Errorf("Error closing database: %v", err)
		} else {
			logx.Info("  ✅ Database connection closed")
		}
	}

	if c.Redis != nil {
		if err := c.Redis.Close(); err != nil {
			logx.Errorf("Error closing Redis: %v", err)
		} else {
			logx.Info("  ✅ Redis connection closed")
		}
	}

	logx.Info("✅ Cleanup complete")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
} // ConsoleNotifier implements the NotificationService interface
// by printing OTP codes to the terminal/console
type ConsoleNotifier struct{}

// NewConsoleNotifier creates a new console-based OTP notifier
func NewConsoleNotifier() *ConsoleNotifier {
	return &ConsoleNotifier{}
}

// SendOTP prints the OTP code to the terminal
func (n *ConsoleNotifier) SendOTP(ctx context.Context, contact string, code string) error {
	fmt.Println("\n" + repeatString("=", 60))
	fmt.Println("📧 OTP NOTIFICATION (Console Output)")
	fmt.Println(repeatString("=", 60))
	fmt.Printf("📨 To: %s\n", contact)
	fmt.Printf("🔐 Code: %s\n", code)
	fmt.Println(repeatString("=", 60))
	fmt.Println("⚠️  This is console output for development only")
	fmt.Println("⚠️  In production, configure email service in config")
	fmt.Println(repeatString("=", 60) + "\n")

	logx.Infof("📧 OTP sent to %s: %s", contact, code)
	return nil
}
