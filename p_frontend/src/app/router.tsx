import {
  Outlet,
  createRootRouteWithContext,
  createRoute,
  createRouter,
  lazyRouteComponent,
  redirect,
} from '@tanstack/react-router'
import { z } from 'zod'
import { AdminLayout } from './layouts/AdminLayout'

type RouterContext = {
  auth: {
    isAuthenticated: boolean
  }
}

const rootRoute = createRootRouteWithContext<RouterContext>()({
  component: Outlet,
})

const loginSearchSchema = z.object({
  redirect: z.string().optional(),
  reason: z.enum(['expired']).optional(),
})

export const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  validateSearch: (search) => loginSearchSchema.parse(search),
  beforeLoad: ({ context, search }) => {
    if (context.auth.isAuthenticated) {
      throw redirect({
        to: search.redirect ?? '/',
      })
    }
  },
  component: lazyRouteComponent(() => import('../pages/LoginPage'), 'LoginPage'),
})

const protectedRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'protected',
  beforeLoad: ({ context, location }) => {
    if (!context.auth.isAuthenticated) {
      throw redirect({
        to: '/login',
        search: {
          redirect: location.href,
        },
      })
    }
  },
  component: AdminLayout,
})

const homeRoute = createRoute({
  getParentRoute: () => protectedRoute,
  path: '/',
  component: lazyRouteComponent(() => import('../pages/HomePage'), 'HomePage'),
})

const organizationStructureRoute = createRoute({
  getParentRoute: () => protectedRoute,
  path: '/organization/departments',
  component: lazyRouteComponent(() => import('../pages/organization/OrganizationStructurePage'), 'OrganizationStructurePage'),
})

const organizationMembersRoute = createRoute({
  getParentRoute: () => protectedRoute,
  path: '/organization/users',
  component: lazyRouteComponent(() => import('../pages/organization/OrganizationMembersPage'), 'OrganizationMembersPage'),
})

const organizationPositionsRoute = createRoute({
  getParentRoute: () => protectedRoute,
  path: '/organization/positions',
  component: lazyRouteComponent(() => import('../pages/organization/OrganizationPositionsPage'), 'OrganizationPositionsPage'),
})

const permissionRolesRoute = createRoute({
  getParentRoute: () => protectedRoute,
  path: '/permission/role-permissions',
  component: lazyRouteComponent(() => import('../pages/permission/PermissionRolesPage'), 'PermissionRolesPage'),
})

const permissionPoliciesRoute = createRoute({
  getParentRoute: () => protectedRoute,
  path: '/permission/menu-permissions',
  component: lazyRouteComponent(() => import('../pages/permission/PermissionPoliciesPage'), 'PermissionPoliciesPage'),
})

const businessUsersRoute = createRoute({
  getParentRoute: () => protectedRoute,
  path: '/business/users',
  component: lazyRouteComponent(() => import('../pages/business/BusinessUsersPage'), 'BusinessUsersPage'),
})

const businessUserPunishmentsRoute = createRoute({
  getParentRoute: () => protectedRoute,
  path: '/business/user-punishments',
  component: lazyRouteComponent(() => import('../pages/business/BusinessUserPunishmentsPage'), 'BusinessUserPunishmentsPage'),
})

const resourceFilesRoute = createRoute({
  getParentRoute: () => protectedRoute,
  path: '/resource/files',
  component: lazyRouteComponent(() => import('../pages/resource/ResourceFilesPage'), 'ResourceFilesPage'),
})

const systemSettingsRoute = createRoute({
  getParentRoute: () => protectedRoute,
  path: '/system/settings',
  component: lazyRouteComponent(() => import('../pages/system/SystemSettingsPage'), 'SystemSettingsPage'),
})

const systemAuditLogsRoute = createRoute({
  getParentRoute: () => protectedRoute,
  path: '/system/audit-logs',
  component: lazyRouteComponent(() => import('../pages/system/SystemAuditLogsPage'), 'SystemAuditLogsPage'),
})

const systemIpBlacklistRoute = createRoute({
  getParentRoute: () => protectedRoute,
  path: '/system/ip-blacklist',
  component: lazyRouteComponent(() => import('../pages/system/SystemIpBlacklistPage'), 'SystemIpBlacklistPage'),
})

const systemWarmTipsRoute = createRoute({
  getParentRoute: () => protectedRoute,
  path: '/system/warm-tips',
  component: lazyRouteComponent(() => import('../pages/system/SystemWarmTipsPage'), 'SystemWarmTipsPage'),
})

const systemAlertBotsRoute = createRoute({
  getParentRoute: () => protectedRoute,
  path: '/system/alert-bots',
  component: lazyRouteComponent(() => import('../pages/system/SystemAlertBotsPage'), 'SystemAlertBotsPage'),
})

const routeTree = rootRoute.addChildren([
  loginRoute,
  protectedRoute.addChildren([
    homeRoute,
    organizationStructureRoute,
    organizationMembersRoute,
    organizationPositionsRoute,
    businessUsersRoute,
    businessUserPunishmentsRoute,
    resourceFilesRoute,
    permissionRolesRoute,
    permissionPoliciesRoute,
    systemSettingsRoute,
    systemAuditLogsRoute,
    systemIpBlacklistRoute,
    systemWarmTipsRoute,
    systemAlertBotsRoute,
  ]),
])

export const router = createRouter({
  routeTree,
  context: {
    auth: {
      isAuthenticated: false,
    },
  },
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
