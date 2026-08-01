// Package grpc предоставляет gRPC-сервер для API сокращения URL.
// Он оборачивает HTTP-сервис и реализует интерфейс gRPC ShortenerService.
package grpc

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/kirillshkro/gshortener/internal/handler/shortener"
	"github.com/kirillshkro/gshortener/internal/model"
	"github.com/kirillshkro/gshortener/internal/proto"
	"github.com/kirillshkro/gshortener/internal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	grpcstatus "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// server реализует интерфейс proto.ShortenerServiceServer
// и делегирует вызовы существующему HTTP-сервису.
type server struct {
	proto.UnimplementedShortenerServiceServer
	service *shortener.Service
	logger  *slog.Logger
}

// NewServer создаёт новый gRPC-сервер поверх существующего HTTP-сервиса.
func NewServer(service *shortener.Service) proto.ShortenerServiceServer {
	return &server{
		service: service,
		logger:  slog.New(slog.NewTextHandler(nil, nil)),
	}
}

// ShortenURL обрабатывает RPC-запрос на сокращение URL.
func (s *server) ShortenURL(ctx context.Context, req *proto.URLShortenRequest) (*proto.URLShortenResponse, error) {
	url := req.GetUrl()
	if url == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "url is required")
	}

	id := shortener.Hashing([]byte(url))
	result := string(s.service.ResultAddr) + "/" + string(id)

	if err := s.service.Stor.Create(model.URLData{
		ShortURL:    types.ShortURL(id),
		OriginalURL: types.RawURL(url),
	}); err != nil {
		var eu *types.ErrUnique
		if errors.As(err, &eu) {
			existing := string(s.service.ResultAddr) + "/" + string(eu.ShortURL)
			return &proto.URLShortenResponse{Result: &existing}, nil
		}
		s.logger.Error("cannot write to storage", "error", err)
		return nil, grpcstatus.Error(codes.Internal, "failed to store URL")
	}

	return &proto.URLShortenResponse{Result: &result}, nil
}

// ExpandURL обрабатывает RPC-запрос на расширение короткого URL.
func (s *server) ExpandURL(ctx context.Context, req *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	id := strings.TrimPrefix(req.GetId(), "/")
	location, err := s.service.Stor.OriginalURL(types.ShortURL(id))
	if err != nil {
		var ad *types.ErrURLDeleted
		if errors.As(err, &ad) {
			return nil, grpcstatus.Error(codes.NotFound, "URL already deleted")
		}
		return nil, grpcstatus.Error(codes.NotFound, "not found")
	}

	return &proto.URLExpandResponse{Result: stringPtr(string(location))}, nil
}

// ListUserURLs обрабатывает RPC-запрос на получение URL пользователя.
// Идентификатор пользователя передаётся в metadata под ключом "user_id".
func (s *server) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*proto.UserURLsResponse, error) {
	var userID string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		userIDs := md.Get("user_id")
		if len(userIDs) > 0 {
			userID = userIDs[0]
		}
	}
	if userID == "" {
		return nil, grpcstatus.Error(codes.Unauthenticated, "user_id is required")
	}

	urls, err := s.service.Stor.GetUserURLs(userID)
	if err != nil {
		return nil, grpcstatus.Error(codes.Internal, "failed to get user URLs")
	}

	var result []*proto.URLData
	for _, u := range urls {
		shortURL := string(s.service.ResultAddr) + "/" + string(u.ShortURL)
		result = append(result, &proto.URLData{
			ShortUrl:    stringPtr(shortURL),
			OriginalUrl: stringPtr(string(u.OriginalURL)),
		})
	}

	return &proto.UserURLsResponse{Url: result}, nil
}

// stringPtr возвращает указатель на строку.
func stringPtr(s string) *string {
	return &s
}
