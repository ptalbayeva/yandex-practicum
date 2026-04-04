package service

import (
	"context"
	"errors"
	"net/url"

	g "github.com/yandex-practicum/shorten-url/internal/middleware"
	"github.com/yandex-practicum/shorten-url/internal/repository"
	"github.com/yandex-practicum/shorten-url/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ShortenerServer struct {
	proto.UnimplementedShortenerServiceServer
	svc *ShortenerService
}

func NewShortenerServer(svc *ShortenerService) *ShortenerServer {
	return &ShortenerServer{svc: svc}
}

func (s *ShortenerServer) ShortenURL(ctx context.Context, req *proto.URLShortenRequest) (*proto.URLShortenResponse, error) {
	userID := s.extractUserID(ctx)

	u, err := s.svc.Shorten(req.GetUrl(), userID)

	if errors.Is(err, repository.ErrConflict) {
		result, _ := url.JoinPath(s.svc.BaseURL, u.Code)
		return &proto.URLShortenResponse{
			Result: result,
		}, status.Error(codes.AlreadyExists, "URL already exists")
	}

	if err != nil {
		g.Log.Error("shorten error", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "shorten error: %v", err)
	}

	result, err := url.JoinPath(s.svc.BaseURL, u.Code)
	if err != nil {
		g.Log.Error("internal path error", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "internal path error")
	}

	return &proto.URLShortenResponse{
		Result: result,
	}, nil
}

func (s *ShortenerServer) ExpandURL(ctx context.Context, req *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	u, err := s.svc.Resolve(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "not found: %v", err)
	}

	if u.IsDeleted {
		return nil, status.Error(codes.FailedPrecondition, "URL was deleted")
	}

	return &proto.URLExpandResponse{
		Result: u.Original,
	}, nil
}

func (s *ShortenerServer) ListUserURLs(ctx context.Context, req *proto.ListUserURLsRequest) (*proto.UserURLsResponse, error) {
	userID := s.extractUserID(ctx)

	urls, err := s.svc.GetManyByUserID(userID)
	if err != nil {
		g.Log.Error("fetch error", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "fetch error: %v", err)
	}

	if len(urls) == 0 {
		return nil, status.Error(codes.NotFound, "no content")
	}

	items := make([]*proto.URLData, 0, len(urls))
	for _, val := range urls {
		items = append(items, &proto.URLData{
			ShortUrl:    val.ShortenURL,
			OriginalUrl: val.OriginalURL,
		})
	}

	return &proto.UserURLsResponse{
		Urls: items,
	}, nil
}

func (s *ShortenerServer) extractUserID(ctx context.Context) string {
	if v := ctx.Value("user_id"); v != nil {
		if userID, ok := v.(string); ok {
			return userID
		}
	}
	return ""
}
