package service

import (
    "context"

    "github.com/manlikehenryy/loan-management-system-grpc/loanService/configs"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
)

func TokenInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return nil, status.Errorf(codes.Unauthenticated, "no metadata provided")
    }

    authHeader := md["authorization"]
    if len(authHeader) == 0 {
        return nil, status.Errorf(codes.Unauthenticated, "no auth token provided")
    }

    token := authHeader[0]
    if len(token) < 7 || token[:7] != "Bearer " {
        return nil, status.Errorf(codes.Unauthenticated, "invalid auth token format")
    }
    actualToken := token[7:]

    if actualToken != configs.Env.TOKEN {
        return nil, status.Errorf(codes.Unauthenticated, "invalid auth token")
    }

    return handler(ctx, req)
}
