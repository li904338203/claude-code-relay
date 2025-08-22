// 模型管理作为系统管理的子页面，不需要独立的顶级菜单
export default [
  {
    path: '/admin/models',
    name: 'AdminModels',
    component: () => import('@/pages/admin/models/index.vue'),
    meta: {
      title: '模型配置',
      hidden: true, // 通过后端动态菜单显示
      keepAlive: true,
    },
  },
];
