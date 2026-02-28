package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/yandex-practicum/shorten-url/internal/model"
	"github.com/yandex-practicum/shorten-url/internal/repository"
	"github.com/yandex-practicum/shorten-url/internal/service"
)

type mockRepo struct {
	repository.MemoryRepo
}

func (m *mockRepo) DeleteManyByCodes(userID string, codes []string) error {
	time.Sleep(1 * time.Millisecond)
	return nil
}

func BenchmarkDeleteURLService(b *testing.B) {
	repo := &mockRepo{}
	s := service.NewDeleteURLService(repo, 1000)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.Run(ctx)

	b.Run("Enqueue_Single", func(b *testing.B) {
		task := model.DeleteURLTask{
			UserID: "user1",
			Short:  "short_code",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s.Enqueue(task)
		}
	})

	b.Run("Enqueue_Parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			task := model.DeleteURLTask{
				UserID: "user_concurrent",
				Short:  "abc",
			}
			for pb.Next() {
				s.Enqueue(task)
			}
		})
	})
}
