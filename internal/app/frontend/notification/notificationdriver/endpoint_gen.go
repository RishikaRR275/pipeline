// Code generated by mga tool. DO NOT EDIT.
package notificationdriver

import (
	"github.com/banzaicloud/pipeline/internal/app/frontend/notification"
	"github.com/go-kit/kit/endpoint"
	kitoc "github.com/go-kit/kit/tracing/opencensus"
	kitxendpoint "github.com/sagikazarmark/kitx/endpoint"
)

// Endpoints collects all of the endpoints that compose the underlying service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
type Endpoints struct {
	GetNotifications endpoint.Endpoint
}

// MakeEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the provided service.
func MakeEndpoints(service notification.Service, middleware ...endpoint.Middleware) Endpoints {
	mw := kitxendpoint.Chain(middleware...)

	return Endpoints{GetNotifications: mw(MakeGetNotificationsEndpoint(service))}
}

// TraceEndpoints returns an Endpoints struct where each endpoint is wrapped with a tracing middleware.
func TraceEndpoints(endpoints Endpoints) Endpoints {
	return Endpoints{GetNotifications: kitoc.TraceEndpoint("notification.GetNotifications")(endpoints.GetNotifications)}
}
