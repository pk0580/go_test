package grpcserver

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/user/go-sender/internal/grpcserver/pb"
	"github.com/user/go-sender/internal/worker"
)

// MockEmailSender для тестов
type MockEmailSender struct {
	mock.Mock
}

func (m *MockEmailSender) Send(recipient, content string) error {
	args := m.Called(recipient, content)
	return args.Error(0)
}

func TestServer_SendEmail(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	
	mockSender := new(MockEmailSender)
	// В данном случае нам не нужен реальный воркер для SendEmail, но он нужен для NewServer
	srv := NewServer(mockSender, &worker.Worker{})
	pb.RegisterSenderServiceServer(s, srv)

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
	defer s.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", 
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewSenderServiceClient(conn)

	t.Run("Success", func(t *testing.T) {
		mockSender.On("Send", "test@example.com", "Hello").Return(nil).Once()

		resp, err := client.SendEmail(ctx, &pb.EmailRequest{
			To:      "test@example.com",
			Subject: "Test",
			Body:    "Hello",
		})

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.MessageId, "gRPC-")
		mockSender.AssertExpectations(t)
	})

	t.Run("Failure", func(t *testing.T) {
		mockSender.On("Send", "fail@example.com", "Bad").Return(assert.AnError).Once()

		resp, err := client.SendEmail(ctx, &pb.EmailRequest{
			To:      "fail@example.com",
			Subject: "Test",
			Body:    "Bad",
		})

		assert.NoError(t, err) // Мы возвращаем ошибку в EmailResponse, а не как ошибку gRPC
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Error, "Ошибка при отправке email")
		mockSender.AssertExpectations(t)
	})
}

func TestServer_GetWorkerStatus(t *testing.T) {
	// Для этого теста нам не нужен сетевой вызов, можем вызвать метод напрямую
	w := &worker.Worker{} // processed по умолчанию 0
	srv := NewServer(nil, w)

	resp, err := srv.GetWorkerStatus(context.Background(), &pb.StatusRequest{})

	assert.NoError(t, err)
	assert.Equal(t, "running", resp.Status)
	assert.Equal(t, int64(0), resp.MessagesProcessed)
}
