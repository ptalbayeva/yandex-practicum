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

	shortURL, joinErr := url.JoinPath(s.svc.BaseURL, u.Code)
	if joinErr != nil {
		g.Log.Error("path error", zap.Error(joinErr))
		return nil, status.Error(codes.Internal, "internal error")
	}

	if errors.Is(err, repository.ErrConflict) {
		return &proto.URLShortenResponse{
			Id:       u.Code,
			ShortUrl: shortURL,
		}, status.Error(codes.AlreadyExists, "already exists")
	}

	if err != nil {
		g.Log.Error("shorten error", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &proto.URLShortenResponse{
		Id:       u.Code,
		ShortUrl: shortURL,
	}, nil
}

func (s *ShortenerServer) ExpandURL(ctx context.Context, req *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	u, err := s.svc.Resolve(req.GetId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}

	if u.IsDeleted {
		return nil, status.Error(codes.FailedPrecondition, "deleted")
	}

	return &proto.URLExpandResponse{
		Url: u.Original,
	}, nil
}

func (s *ShortenerServer) ListUserURLs(ctx context.Context, req *proto.ListUserURLsRequest) (*proto.UserURLsResponse, error) {
	userID := s.extractUserID(ctx)

	urls, err := s.svc.GetManyByUserID(userID)
	if err != nil {
		g.Log.Error("fetch error", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	if len(urls) == 0 {
		return nil, status.Error(codes.NotFound, "no content")
	}

	items := make([]*proto.URLRef, 0, len(urls))
	for _, val := range urls {
		shortURL, err := url.JoinPath(s.svc.BaseURL, val.ShortenURL)
		if err != nil {
			g.Log.Error("path error", zap.Error(err))
			continue
		}

		items = append(items, &proto.URLRef{
			Id:       val.ShortenURL,
			ShortUrl: shortURL,
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
