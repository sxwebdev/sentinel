package apiserver

import (
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"github.com/sxwebdev/rbacconnect"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/agents/v1/agentsv1connect"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/auth/v1/authv1connect"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/notifications/v1/notificationsv1connect"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/system/v1/systemv1connect"
	"github.com/sxwebdev/sentinel/internal/models"
)

const (
	UserRoleAnonymous rbacconnect.Role = "anon"
	UserRoleSetup     rbacconnect.Role = "setup"
)

func UserPolicy() *rbacconnect.Policy {
	b := rbacconnect.NewPolicyBuilder().
		WithSuperRoles(models.UserRoleRoot).
		WithDefaultAllow(true)

	// Deny anonymous users to access setup procedures
	b.When(rbacconnect.Any()).Deny(UserRoleAnonymous)

	allowAnonServices := []string{
		grpcreflect.ReflectV1AlphaServiceName, // reflection
		grpcreflect.ReflectV1ServiceName,
		grpchealth.HealthV1ServiceName, // health
	}
	rbacconnect.AllowServices(b, UserRoleAnonymous, allowAnonServices...)

	allowAnonProcs := []string{
		authv1connect.AuthServiceAuthorizationProcedure,
		authv1connect.AuthServiceRefreshTokenProcedure,
		systemv1connect.SystemServiceCheckIsInitializedProcedure,
		systemv1connect.SystemServiceInitializeProcedure,
	}
	rbacconnect.AllowProcs(b, UserRoleAnonymous, allowAnonProcs...)

	// Deny users in setup mode to access other procedures
	b.When(rbacconnect.Any()).Deny(UserRoleSetup)

	rbacconnect.AllowProcs(b, UserRoleSetup,
		systemv1connect.SystemServiceCheckIsInitializedProcedure,
		systemv1connect.SystemServiceInitializeProcedure,
	)

	// Deny regular users to access admin-only procedures
	rbacconnect.DenyProcs(b, models.UserRoleUser,
		notificationsv1connect.NotificationServiceHistoryDeleteAllProcedure,
		notificationsv1connect.NotificationServiceProviderUpdateProcedure,
		notificationsv1connect.NotificationServiceProviderDeleteProcedure,
		agentsv1connect.AgentsServiceAgentsDeleteProcedure,
		agentsv1connect.AgentsServiceAgentsUpdateProcedure,
	)

	return b.Build()
}
