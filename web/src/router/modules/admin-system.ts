// 系统管理相关路由配置 - 管理员页面
export default [
  {
    path: '/admin/accounts',
    name: 'AdminAccounts',
    component: () => import('@/pages/admin/accounts/index.vue'),
    meta: {
      title: '账号管理',
      hidden: true, // 通过后端动态菜单显示
      keepAlive: true,
      roles: ['admin'],
    },
  },
  {
    path: '/admin/keys',
    name: 'AdminKeys',
    component: () => import('@/pages/admin/keys/index.vue'),
    meta: {
      title: '密钥管理',
      hidden: true, // 通过后端动态菜单显示
      keepAlive: true,
      roles: ['admin'],
    },
  },
  {
    path: '/admin/groups',
    name: 'AdminGroups',
    component: () => import('@/pages/admin/groups/index.vue'),
    meta: {
      title: '分组管理',
      hidden: true, // 通过后端动态菜单显示
      keepAlive: true,
      roles: ['admin'],
    },
  },
  {
    path: '/admin/logs/all',
    name: 'AdminAllLogs',
    component: () => import('@/pages/admin/logs/all/index.vue'),
    meta: {
      title: '所有日志',
      hidden: true, // 通过后端动态菜单显示
      keepAlive: true,
      roles: ['admin'],
    },
  },
  {
    path: '/admin/invite',
    name: 'AdminInvite',
    component: () => import('@/pages/admin/invite/index.vue'),
    meta: {
      title: '邀请管理',
      hidden: true, // 通过后端动态菜单显示
      keepAlive: true,
      roles: ['admin'],
    },
  },
];
