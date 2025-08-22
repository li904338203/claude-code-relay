export default [
  {
    path: '/admin/billing/plans',
    name: 'AdminBillingPlans',
    component: () => import('@/pages/admin/billing/plans/index.vue'),
    meta: {
      title: '用户套餐管理',
      hidden: true,
    },
  },
];
