package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"vitract-assessment/internal/database"
	"vitract-assessment/internal/notification"
	"vitract-assessment/internal/patient"
	"vitract-assessment/internal/sample"
	transportHTTP "vitract-assessment/internal/transport/http"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Println("Successfully connected to PostgreSQL")

	patientRepo := patient.NewRepository(conn)
	importer := patient.NewImporter(patientRepo)

	err = importer.ImportFromCSV(ctx, "../data/patients.csv")
	if err != nil {
		log.Printf("Warning during initial CSV import: %v", err)
	} else {
		log.Println("Patients roster verified/imported successfully from CSV")
	}

	sampleRepo := sample.NewRepository(conn)
	processor := sample.NewProcessor(sampleRepo, patientRepo)
	emailSender := notification.NewEmailSender()

	jobQueue := make(chan sample.Job, 100)

	worker := sample.NewWorker(jobQueue, processor, emailSender)
	go worker.Start(ctx)

	samplesDir := "../data/samples"
	log.Printf("Scanning directory '%s' for lab results...", samplesDir)
	
	files, err := os.ReadDir(samplesDir)
	if err != nil {
		log.Printf("Warning: could not read samples directory: %v", err)
	} else {
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
				continue
			}

			sampleID := strings.TrimSuffix(file.Name(), ".json")
			filePath := filepath.Join(samplesDir, file.Name())

			log.Printf("Sample %s targets ingestion pipeline...", sampleID)
			
			jobQueue <- sample.Job{
				SampleID: sampleID,
				FilePath: filePath,
			}
		}
	}

	mux := http.NewServeMux()
	handler := transportHTTP.NewHandler(patientRepo, sampleRepo, jobQueue)
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("HTTP Server is running on http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server gracefully...")
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly. Goodbye!")
}