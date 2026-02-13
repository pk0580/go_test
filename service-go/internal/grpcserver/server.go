package grpcserver

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/user/go-sender/internal/grpcserver/pb"
	"github.com/user/go-sender/internal/sender"
)

// Server реализует gRPC сервис SenderService
type Server struct {
	pb.UnimplementedSenderServiceServer
	emailSender sender.EmailSender
	processed   uint64 // Количество обработанных сообщений через gRPC
}

// NewServer создает новый экземпляр gRPC сервера
func NewServer(s sender.EmailSender) *Server {
	return &Server{
		emailSender: s,
	}
}

// SendEmail обрабатывает запрос на отправку письма
func (s *Server) SendEmail(ctx context.Context, req *pb.EmailRequest) (*pb.EmailResponse, error) {
	log.Info().
		Str("recipient", req.GetTo()).
		Str("subject", req.GetSubject()).
		Msg("gRPC: Received SendEmail request")

	err := s.emailSender.Send(req.GetTo(), req.GetBody())
	if err != nil {
		log.Error().Err(err).Msg("gRPC: Ошибка при отправке email")
		return &pb.EmailResponse{
			Success: false,
			Error:   fmt.Sprintf("Ошибка при отправке email: %v", err),
		}, nil
	}

	atomic.AddUint64(&s.processed, 1)
	log.Info().Msg("gRPC: Email успешно отправлен")

	return &pb.EmailResponse{
		Success: true,
		MessageId: "gRPC-" + fmt.Sprint(time.Now().UnixNano()),
	}, nil
}

// GetWorkerStatus возвращает статус воркера
func (s *Server) GetWorkerStatus(ctx context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	processed := atomic.LoadUint64(&s.processed)
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
