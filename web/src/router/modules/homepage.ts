import { DashboardIcon } from 'tdesign-icons-vue-next';
import { shallowRef } from 'vue';

import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/dashboard',
    component: Layout,
    redirect: '/dashboard/base',
    name: 'dashboard',
    meta: {
      title: '仪表盘',
      icon: shallowRef(DashboardIcon),
      orderNo: 0,
      hidden: true, // 隐藏静态菜单，完全使用后端动态菜单
    },
    children: [
      {
        path: 'base',
        name: 'DashboardBase',
        component: () => import('@/pages/dashboard/base/index.vue'),
        meta: {
          title: '概览仪表盘',
        },
      },
    ],
  },
];
