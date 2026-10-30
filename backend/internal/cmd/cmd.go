package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"backend/internal/controller/hello"
	"backend/internal/controller/tryon"
	"backend/internal/imagevalidate"
	"backend/internal/middleware"
	"backend/internal/security"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			secService, err := security.InitService(ctx)
			if err != nil {
				return err
			}

			imgLimits := imagevalidate.LoadImageLimits(ctx)
			corsMw := middleware.NewCORSMiddleware(secService.Config)
			bodyLimitMw := middleware.NewBodyLimitMiddleware(imgLimits.MaxRequestBodyBytes)
			authMw := middleware.NewAuthMiddleware(secService.TokenStore)
			rateLimitMw := middleware.NewRateLimitMiddleware(secService.RateLimiter)

			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(
					middleware.RequestID,
					ghttp.MiddlewareHandlerResponse,
					corsMw.Handler,
					bodyLimitMw.Handler,
					authMw.Handler,
					rateLimitMw.Handler,
				)
				group.Bind(
					hello.NewV1(),
					tryon.NewV1(),
				)
			})
			s.Run()
			return nil
		},
	}
)
