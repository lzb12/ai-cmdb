import { createRouter, createWebHistory } from "vue-router";
import MainLayout from "@/layouts/MainLayout.vue";
import LoginView from "@/views/LoginView.vue";
import HostAssetsView from "@/views/HostAssetsView.vue";
import CredentialsView from "@/views/CredentialsView.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/login",
      name: "login",
      meta: { public: true, title: "登录" },
      component: LoginView
    },
    {
      path: "/",
      component: MainLayout,
      children: [
        {
          path: "",
          redirect: "/assets"
        },
        {
          path: "/assets",
          name: "assets",
          meta: { title: "主机资产" },
          component: HostAssetsView
        },
        {
          path: "/credentials",
          name: "credentials",
          meta: { title: "凭据管理" },
          component: CredentialsView
        }
      ]
    }
  ]
});

router.beforeEach((to) => {
  const hasToken = Boolean(localStorage.getItem("cmdb_token"));
  if (!to.meta.public && !hasToken) {
    return { path: "/login" };
  }
  if (to.path === "/login" && hasToken) {
    return { path: "/assets" };
  }
  return true;
});

export default router;
