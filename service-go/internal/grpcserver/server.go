package grpcserver

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/user/go-sender/internal/grpcserver/pb"
	"github.com/user/go-sender/internal/metrics"
	"github.com/user/go-sender/internal/sender"
	"github.com/user/go-sender/internal/worker"
)

// Server реализует gRPC сервис SenderService
type Server struct {
	pb.UnimplementedSenderServiceServer
	emailSender sender.EmailSender
	worker      *worker.Worker
}

// NewServer создает новый экземпляр gRPC сервера
func NewServer(s sender.EmailSender, w *worker.Worker) *Server {
	return &Server{
		emailSender: s,
		worker:      w,
	}
}

// SendEmail обрабатывает запрос на отправку письма
func (s *Server) SendEmail(ctx context.Context, req *pb.EmailRequest) (*pb.EmailResponse, error) {
	start := time.Now()
	log.Info().
		Str("recipient", req.GetTo()).
		Str("subject", req.GetSubject()).
		Msg("gRPC: Received SendEmail request")

	err := s.emailSender.Send(req.GetTo(), req.GetBody())
	
	status := "sent"
	if err != nil {
		status = "failed"
		log.Error().Err(err).Msg("gRPC: Ошибка при отправке email")
		
		duration := time.Since(start).Seconds()
		metrics.MessagesProcessingDuration.Observe(duration)
		metrics.MessagesProcessed.WithLabelValues("grpc_" + status).Inc()

		return &pb.EmailResponse{
			Success: false,
			Error:   fmt.Sprintf("Ошибка при отправке email: %v", err),
		}, nil
	}

	duration := time.Since(start).Seconds()
	metrics.MessagesProcessingDuration.Observe(duration)
	metrics.MessagesProcessed.WithLabelValues("grpc_" + status).Inc()
	
	log.Info().Msg("gRPC: Email успешно отправлен")

	return &pb.EmailResponse{
		Success: true,
		MessageId: "gRPC-" + fmt.Sprint(time.Now().UnixNano()),
	}, nil
}

// GetWorkerStatus возвращает статус воркера
func (s *Server) GetWorkerStatus(ctx context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	processed := s.worker.GetProcessedCount()
	return &pb.StatusResponse{
		Status:            "running",
		MessagesProcessed: int64(processed),
	}, nil
}

// Start запускает gRPC сервер
func (s *Server) Start(ctx context.Context, wg *sync.WaitGroup, port int) {
	defer wg.Done()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal().Err(err).Msg("gRPC: Failed to listen")
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSenderServiceServer(grpcServer, s)

	log.Info().Int("port", port).Msg("gRPC сервер запускается")

	// Канал для ошибок запуска сервера
	serverErr := make(chan error, 1)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			serverErr <- err
		}
	}()

	// Ожидание завершения
	select {
	case <-ctx.Done():
		log.Info().Msg("gRPC сервер останавливается...")
		grpcServer.GracefulStop()
	case err := <-serverErr:
		log.Error().Err(err).Msg("gRPC сервер вернул ошибку")
	}
}
