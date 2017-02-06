package ginpprof

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

// Wrap adds several routes from package `net/http/pprof` to *gin.Engine object
func Wrap(router *gin.Engine) {
	router.GET("/debug/pprof/", IndexHandler())
	router.GET("/debug/pprof/heap", HeapHandler())
	router.GET("/debug/pprof/goroutine", GoroutineHandler())
	router.GET("/debug/pprof/block", BlockHandler())
	router.GET("/debug/pprof/threadcreate", ThreadCreateHandler())
	router.GET("/debug/pprof/cmdline", CmdlineHandler())
	router.GET("/debug/pprof/profile", ProfileHandler())
	router.GET("/debug/pprof/symbol", SymbolHandler())
}

// make sure we are backward compatible
var Wrapper = Wrap

func IndexHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pprof.Index(ctx.Writer, ctx.Request)
	}
}

func HeapHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pprof.Handler("heap").ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func GoroutineHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pprof.Handler("goroutine").ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func BlockHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pprof.Handler("block").ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func ThreadCreateHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pprof.Handler("threadcreate").ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func CmdlineHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pprof.Cmdline(ctx.Writer, ctx.Request)
	}
}

func ProfileHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pprof.Profile(ctx.Writer, ctx.Request)
	}
}

func SymbolHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		pprof.Symbol(ctx.Writer, ctx.Request)
	}
}
