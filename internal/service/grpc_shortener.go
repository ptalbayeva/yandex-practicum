package service

import (
	"context"
	"errors"

	"github.com/yandex-practicum/shorten-url/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/yandex-practicum/shorten-url/internal/repository"
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

	u, err := s.svc.Shorten(req.Url, userID)

	if errors.Is(err, repository.ErrConflict) {
		return &proto.URLShortenResponse{
			Result: s.svc.BaseURL + "/" + u.Code,
		}, status.Error(codes.AlreadyExists, "URL already exists")
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "shorten error: %v", err)
	}

	return &proto.URLShortenResponse{
		Result: s.svc.BaseURL + "/" + u.Code,
	}, nil
}

func (s *ShortenerServer) ExpandURL(ctx context.Context, req *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	u, err := s.svc.Resolve(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "not found: %v", err)
	}

	if u.IsDeleted {
		return nil, status.Error(codes.FailedPrecondition, "URL was deleted by user")
	}

	return &proto.URLExpandResponse{Result: u.Original}, nil
}

func (s *ShortenerServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*proto.UserURLsResponse, error) {
	userID := s.extractUserID(ctx)

	urls, err := s.svc.GetManyByUserID(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch error: %v", err)
	}

	if len(urls) == 0 {
		return nil, status.Error(codes.NotFound, "no content")
	}

	resp := &proto.UserURLsResponse{}
	for _, val := range urls {
		resp.Url = append(resp.Url, &proto.URLData{
			ShortUrl:    val.ShortenURL,
			OriginalUrl: val.OriginalURL,
		})
	}
	return resp, nil
}

func (s *ShortenerServer) extractUserID(ctx context.Context) string {
	if v := ctx.Value("user_id"); v != nil {
		return v.(string)
	}
	return ""
}
