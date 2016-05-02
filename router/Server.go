package router

import (
	"github.com/gocraft/web"
	"net/http"
	"github.com/sandstorm/mailer-daemon/recipientsRepository"
)

type Server struct {
	AuthToken string
	Repository recipientsRepository.Repository
	ServerConfiguration ServerConfiguration
}

func (this *Server) Listen() error {
	if len(this.AuthToken) <= 0 {
		return &ServerError{"AUTH_TOKEN must not be empty."}
	}

	router := web.New(RequestHandler{
		/* handler cannot be initialized here, default values are ignored */
	})
	router.Middleware(web.LoggerMiddleware)
	router.Middleware(web.ShowErrorsMiddleware)
	// to initialize the handler we have to add a middleware where we can access this
	router.Middleware(func(h *RequestHandler, rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
		this.initializeHandler(h, rw, req, next)
	})
	router.Get("/" + this.AuthToken + "/newsletter/status", (*RequestHandler).HandleStatus)
	router.Get("/" + this.AuthToken + "/newsletter/sendingFailures", (*RequestHandler).HandleSendingFailures)
	router.Get("/" + this.AuthToken + "/newsletter/serverConfiguration", (*RequestHandler).HandleServerConfiguration)
	router.Post("/" + this.AuthToken + "/newsletter/:id/send", (*RequestHandler).HandleSend)
	router.Delete("/" + this.AuthToken + "/newsletter/:id/abortAndRemove", (*RequestHandler).HandleAbortAndRemove)
	http.ListenAndServe("localhost:3000", router)
	return nil
}

func (this *Server) initializeHandler(h *RequestHandler, rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	h.Repository = this.Repository
	h.ServerConfiguration = this.ServerConfiguration
	next(rw, req)
}