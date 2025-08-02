package main

import (
	"hotelsdatapipeline/application"
	"hotelsdatapipeline/infra"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("Starting Hotels Data Pipeline...")
	config, err := infra.LoadConfig("config/test.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Println("Configuration loaded successfully")
	redisRepo, err := infra.NewRedisRepository(config.Redis.Host, config.Redis.Port, config.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to initialize Redis repository: %v", err)
	}
	defer redisRepo.Close()
	log.Println("Redis repository initialized successfully")
	hotelFetcher := application.NewHotelFetcher(redisRepo, config.Hotels.URLs)
	log.Println("Hotel fetcher service created")
	cronService := application.NewCronJobService(hotelFetcher, config.CronJob.Interval)
	log.Println("Cron job service created")
	httpServer := application.NewHTTPServer(config.HTTP.Host, config.HTTP.Port, redisRepo)
	log.Printf("HTTP server created on %s", httpServer.GetAddress())
	log.Println("Running initial hotel data fetch...")
	if err := hotelFetcher.FetchAndProcess(); err != nil {
		log.Printf("Initial hotel fetch failed: %v", err)
	} else {
		log.Println("Initial hotel fetch completed successfully")
	}
	if err := cronService.Start(); err != nil {
		log.Fatalf("Failed to start cron service: %v", err)
	}
	log.Println("Cron scheduler started")
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()
	log.Printf("Hotels Data Pipeline is running on %s", httpServer.GetAddress())
	log.Println("Press Ctrl+C to stop the application")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutdown signal received, starting graceful shutdown...")
	if err := httpServer.Stop(); err != nil {
		log.Printf("Error stopping HTTP server: %v", err)
	}
	cronService.Stop()
	if err := redisRepo.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}
	log.Println("Hotels Data Pipeline shutdown completed")
}
