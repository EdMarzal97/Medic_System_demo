package sample

import (
	"context"
	"log"
	"vitract-assessment/internal/notification"
)

type Job struct {
	SampleID string
	FilePath string
}

type Worker struct {
	queue       chan Job
	processor   *Processor
	emailSender *notification.EmailSender
}

func NewWorker(queue chan Job, processor *Processor, emailSender *notification.EmailSender) *Worker {
	return &Worker{
		queue:       queue,
		processor:   processor,
		emailSender: emailSender,
	}
}

func (w *Worker) Start(ctx context.Context) {
	log.Println("Background lab ingestion worker started successfully...")
	for {
		select {
		case job := <-w.queue:
			log.Printf("[Worker] Starting processing for sample %s", job.SampleID)

			actualSampleID, err := w.processor.ProcessFile(ctx, job.FilePath, job.SampleID)
			if err != nil {
				log.Printf("[Worker] Error processing sample %s: %v", job.SampleID, err)
				continue
			}

			log.Printf("[Worker] Sample %s processed successfully", actualSampleID)

			fullData, err := w.processor.repository.GetFullResult(ctx, actualSampleID)
			if err == nil {
				pData, err := w.processor.patientRepo.GetByPatientID(ctx, fullData.PatientID)
				if err == nil {
					err = w.emailSender.NotifyPractitioner(pData.PractitionerEmail, actualSampleID, pData.Name)
					if err != nil {
						log.Printf("[Worker] Failed to send email notification for %s: %v", actualSampleID, err)
					}
				}
			}

		case <-ctx.Done():
			log.Println("Shutting down background worker...")
			return
		}
	}
}