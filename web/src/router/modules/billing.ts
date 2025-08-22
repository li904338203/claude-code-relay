import { RouteRecordRaw } from 'vue-router';
import Layout from '@/layouts/index.vue';

const billingRoutes: RouteRecordRaw[] = [
  // 用户计费功能（普通用户和管理员都可访问）
  {
    path: '/billing',
    name: 'billing',
    component: Layout,
    redirect: '/billing/balance',
    meta: {
      title: '钱包服务',
      icon: 'money-circle',
      orderNo: 5,
      hidden: true, // 隐藏静态菜单，完全使用后端动态菜单
    },
    children: [
      {
        path: 'balance',
        name: 'BillingBalance',
        component: () => import('@/pages/billing/balance/index.vue'),
        meta: {
          title: '余额管理',
          icon: 'wallet',
        },
      },
      {
        path: 'consumption',
        name: 'BillingConsumption',
        component: () => import('@/pages/billing/consumption/index.vue'),
        meta: {
          title: '消费历史',
          icon: 'chart-line',
        },
      },
    ],
  },
  // 管理员计费功能（仅管理员可访问）
  {
    path: '/admin/billing',
    name: 'adminBilling',
    component: Layout,
    redirect: '/admin/billing/cards',
    meta: {
      title: '管理员计费',
      icon: 'money-circle',
      orderNo: 8,
      roles: ['admin'], // 仅管理员可访问
      hidden: true, // 在动态菜单中隐藏，通过后端菜单API控制
    },
    children: [
      {
        path: 'cards',
        name: 'AdminBillingCards',
        component: () => import('@/pages/admin/billing/cards/index.vue'),
        meta: {
          title: '充值卡管理',
          icon: 'creditcard',
          roles: ['admin'],
        },
      },
      {
        path: 'stats',
        name: 'AdminBillingStats',
        component: () => import('@/pages/admin/billing/stats/index.vue'),
        meta: {
          title: '消费统计',
          icon: 'chart-bar',
          roles: ['admin'],
        },
      },
    ],
  },
];

export default billingRoutes;
