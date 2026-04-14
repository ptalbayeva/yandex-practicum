package g

import (
	"context"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthGRPCInterceptor — авторизация для grpc
func AuthGRPCInterceptor(key []byte) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var userID string
		var token string

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get(cookieName)
			if len(values) > 0 {
				token = values[0]
			}
		}

		if token != "" {
			if claims, err := parseToken(token, key); err == nil && claims.UserID != "" {
				userID = claims.UserID
			}
		}

		if userID == "" {
			userID = uuid.NewString()
			newToken, err := createToken(userID, key)
			if err == nil {
				header := metadata.Pairs(cookieName, newToken)
				grpc.SendHeader(ctx, header)
			}
		}

		newCtx := context.WithValue(ctx, userIDContextKey, userID)
		log.Println(newCtx.Value(userIDContextKey))

		return handler(newCtx, req)
	}
}
