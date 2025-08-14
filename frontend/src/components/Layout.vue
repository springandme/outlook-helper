<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '200px'" class="sidebar">
      <div class="logo">
        <div v-if="!isCollapse" class="logo-full">
          <svg class="logo-icon" viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg">
            <path d="M896 42.666667H341.333333c-25.6 0-42.666667 17.066667-42.666666 42.666666v42.666667l341.333333 106.666667L938.666667 128V85.333333c0-25.6-17.066667-42.666667-42.666667-42.666666z" fill="#0364B8"></path>
            <path d="M1024 507.733333c0-8.533333-4.266667-17.066667-12.8-21.333333l-366.933333-209.066667s-4.266667 0-4.266667-4.266666c-8.533333-4.266667-12.8-4.266667-21.333333-4.266667s-17.066667 0-21.333334 4.266667c0 0-4.266667 0-4.266666 4.266666l-366.933334 209.066667c-8.533333 4.266667-12.8 12.8-12.8 21.333333s4.266667 17.066667 12.8 21.333334l366.933334 209.066666s4.266667 0 4.266666 4.266667c8.533333 4.266667 12.8 4.266667 21.333334 4.266667s17.066667 0 21.333333-4.266667c0 0 4.266667 0 4.266667-4.266667l366.933333-209.066666c8.533333-4.266667 12.8-12.8 12.8-21.333334z" fill="#0A2767"></path>
            <path d="M640 128H298.666667v298.666667l341.333333 341.333333h298.666667v-341.333333z" fill="#28A8EA"></path>
            <path d="M640 128h298.666667v298.666667h-298.666667z" fill="#50D9FF"></path>
            <path d="M298.666667 426.666667h341.333333v426.666666H298.666667z" fill="#0078D4"></path>
            <path d="M546.133333 810.666667H51.2C21.333333 810.666667 0 789.333333 0 759.466667V264.533333C0 234.666667 21.333333 213.333333 51.2 213.333333h499.2c25.6 0 46.933333 21.333333 46.933333 51.2v499.2c0 25.6-21.333333 46.933333-51.2 46.933334z" fill="#0078D4"></path>
            <path d="M157.866667 426.666667c12.8-25.6 29.866667-46.933333 55.466666-64 25.6-12.8 55.466667-21.333333 89.6-21.333334 29.866667 0 59.733333 8.533333 81.066667 21.333334 25.6 12.8 42.666667 34.133333 55.466667 59.733333 12.8 25.6 17.066667 55.466667 17.066666 85.333333 0 34.133333-8.533333 64-21.333333 89.6-8.533333 29.866667-29.866667 51.2-51.2 64-25.6 12.8-55.466667 21.333333-85.333333 21.333334-34.133333 0-59.733333-8.533333-85.333334-21.333334-25.6-12.8-42.666667-34.133333-55.466666-59.733333-12.8-25.6-21.333333-55.466667-21.333334-85.333333 0-34.133333 8.533333-64 21.333334-89.6z m59.733333 145.066666c8.533333 17.066667 17.066667 29.866667 29.866667 42.666667 12.8 8.533333 29.866667 12.8 51.2 12.8s38.4-4.266667 51.2-17.066667c12.8-8.533333 25.6-25.6 29.866666-42.666666 8.533333-12.8 12.8-34.133333 12.8-55.466667s-4.266667-38.4-8.533333-55.466667-17.066667-29.866667-29.866667-42.666666c-17.066667-12.8-34.133333-17.066667-55.466666-17.066667s-38.4 4.266667-51.2 17.066667c-12.8 8.533333-25.6 25.6-34.133334 42.666666-4.266667 12.8-8.533333 34.133333-8.533333 55.466667s4.266667 42.666667 12.8 59.733333z" fill="#FFFFFF"></path>
          </svg>
          <h2>{{ appTitle }}</h2>
        </div>
        <div v-else class="logo-collapsed">
          <svg class="logo-icon-collapsed" viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg">
            <path d="M896 42.666667H341.333333c-25.6 0-42.666667 17.066667-42.666666 42.666666v42.666667l341.333333 106.666667L938.666667 128V85.333333c0-25.6-17.066667-42.666667-42.666667-42.666666z" fill="#0364B8"></path>
            <path d="M1024 507.733333c0-8.533333-4.266667-17.066667-12.8-21.333333l-366.933333-209.066667s-4.266667 0-4.266667-4.266666c-8.533333-4.266667-12.8-4.266667-21.333333-4.266667s-17.066667 0-21.333334 4.266667c0 0-4.266667 0-4.266666 4.266666l-366.933334 209.066667c-8.533333 4.266667-12.8 12.8-12.8 21.333333s4.266667 17.066667 12.8 21.333334l366.933334 209.066666s4.266667 0 4.266666 4.266667c8.533333 4.266667 12.8 4.266667 21.333334 4.266667s17.066667 0 21.333333-4.266667c0 0 4.266667 0 4.266667-4.266667l366.933333-209.066666c8.533333-4.266667 12.8-12.8 12.8-21.333334z" fill="#0A2767"></path>
            <path d="M640 128H298.666667v298.666667l341.333333 341.333333h298.666667v-341.333333z" fill="#28A8EA"></path>
            <path d="M640 128h298.666667v298.666667h-298.666667z" fill="#50D9FF"></path>
            <path d="M298.666667 426.666667h341.333333v426.666666H298.666667z" fill="#0078D4"></path>
            <path d="M546.133333 810.666667H51.2C21.333333 810.666667 0 789.333333 0 759.466667V264.533333C0 234.666667 21.333333 213.333333 51.2 213.333333h499.2c25.6 0 46.933333 21.333333 46.933333 51.2v499.2c0 25.6-21.333333 46.933333-51.2 46.933334z" fill="#0078D4"></path>
            <path d="M157.866667 426.666667c12.8-25.6 29.866667-46.933333 55.466666-64 25.6-12.8 55.466667-21.333333 89.6-21.333334 29.866667 0 59.733333 8.533333 81.066667 21.333334 25.6 12.8 42.666667 34.133333 55.466667 59.733333 12.8 25.6 17.066667 55.466667 17.066666 85.333333 0 34.133333-8.533333 64-21.333333 89.6-8.533333 29.866667-29.866667 51.2-51.2 64-25.6 12.8-55.466667 21.333333-85.333333 21.333334-34.133333 0-59.733333-8.533333-85.333334-21.333334-25.6-12.8-42.666667-34.133333-55.466666-59.733333-12.8-25.6-21.333333-55.466667-21.333334-85.333333 0-34.133333 8.533333-64 21.333334-89.6z m59.733333 145.066666c8.533333 17.066667 17.066667 29.866667 29.866667 42.666667 12.8 8.533333 29.866667 12.8 51.2 12.8s38.4-4.266667 51.2-17.066667c12.8-8.533333 25.6-25.6 29.866666-42.666666 8.533333-12.8 12.8-34.133333 12.8-55.466667s-4.266667-38.4-8.533333-55.466667-17.066667-29.866667-29.866667-42.666666c-17.066667-12.8-34.133333-17.066667-55.466666-17.066667s-38.4 4.266667-51.2 17.066667c-12.8 8.533333-25.6 25.6-34.133334 42.666666-4.266667 12.8-8.533333 34.133333-8.533333 55.466667s4.266667 42.666667 12.8 59.733333z" fill="#FFFFFF"></path>
          </svg>
        </div>
      </div>

      <el-menu
        :default-active="activeMenu"
        class="sidebar-menu"
        :collapse="isCollapse"
        background-color="#2D3748"
        text-color="#E2E8F0"
        active-text-color="#4299E1"
        router
      >
        <el-menu-item index="/">
          <el-icon><DataBoard /></el-icon>
          <template #title>仪表盘</template>
        </el-menu-item>

        <el-menu-item index="/emails">
          <el-icon><Message /></el-icon>
          <template #title>邮箱管理</template>
        </el-menu-item>

        <el-menu-item index="/tags">
          <el-icon><Collection /></el-icon>
          <template #title>标签管理</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 主内容区 -->
    <el-container>
      <!-- 顶部导航 -->
      <el-header class="header">
        <div class="header-left">
          <el-button
            type="text"
            @click="toggleCollapse"
            class="collapse-btn"
          >
            <el-icon size="18">
              <Expand v-if="isCollapse" />
              <Fold v-else />
            </el-icon>
          </el-button>

          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="breadcrumbTitle">{{ breadcrumbTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-dropdown">
              <el-icon><User /></el-icon>
              <span>{{ authStore.user?.username }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 主内容 -->
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import {
  DataBoard,
  Message,
  Collection,
  User,
  ArrowDown,
  SwitchButton,
  Expand,
  Fold
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const appTitle = import.meta.env.VITE_APP_TITLE || 'Outlook取件助手'
const isCollapse = ref(false)

const activeMenu = computed(() => route.path)

const breadcrumbTitle = computed(() => {
  const routeMap: Record<string, string> = {
    '/': '仪表盘',
    '/emails': '邮箱管理',
    '/tags': '标签管理'
  }
  return routeMap[route.path] || ''
})

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

const handleCommand = async (command: string) => {
  if (command === 'logout') {
    try {
      await ElMessageBox.confirm(
        '确定要退出登录吗？',
        '提示',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )

      await authStore.logout()
      router.push('/login')
    } catch (error) {
      // 用户取消操作
    }
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: #2D3748;
  transition: width 0.3s;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #4A5568;
  margin-bottom: 10px;
  overflow: hidden; /* 防止文字溢出 */
}

.logo-full {
  display: flex;
  align-items: center;
  gap: 8px;
}

.logo-collapsed {
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.logo-icon-collapsed {
  width: 24px;
  height: 24px;
}

.logo h2 {
  color: #E2E8F0;
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  white-space: nowrap; /* 防止文字换行 */
  transition: opacity 0.3s ease; /* 添加过渡效果 */
}

.sidebar-menu {
  border: none;
}

/* 菜单折叠时图标居中 */
:deep(.el-menu--collapse) .el-menu-item,
:deep(.el-menu--collapse) .el-sub-menu > .el-sub-menu__title {
  display: flex;
  justify-content: center;
  align-items: center;
  padding-left: 0 !important;
  padding-right: 0 !important;
}

/* 修复tooltip wrapper导致的图标不居中问题 */
:deep(.el-menu--collapse) .el-menu-item .el-menu-tooltip__trigger,
:deep(.el-menu--collapse) .el-menu-item .el-tooltip__trigger {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100%;
}

:deep(.el-menu--collapse) .el-menu-item .el-icon {
  margin: 0 !important;
}
:deep(.el-menu--collapse) .el-menu-item span {
  display: none;
}

/* 激活状态的菜单项图标居中 */
:deep(.el-menu--collapse) .el-menu-item.is-active {
  justify-content: center !important;
  align-items: center !important;
}
:deep(.el-menu--collapse) .el-menu-item.is-active .el-menu-tooltip__trigger,
:deep(.el-menu--collapse) .el-menu-item.is-active .el-tooltip__trigger {
  display: flex;
  justify-content: center !important;
  align-items: center !important;
  width: 100%;
  height: 100%;
}
:deep(.el-menu--collapse) .el-menu-item.is-active .el-icon {
  margin: 0 !important;
}
:deep(.el-menu--collapse) .el-sub-menu.is-active > .el-sub-menu__title {
  justify-content: center !important;
  align-items: center !important;
}


.header {
  background-color: #FFFFFF;
  border-bottom: 1px solid #E2E8F0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.collapse-btn {
  color: #4A5568;
  font-size: 18px;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #4A5568;
  font-size: 14px;
}

.user-dropdown:hover {
  color: #2D3748;
}

.main-content {
  background-color: #f5f5f5;
  background-image:
    linear-gradient(90deg, rgba(200, 200, 200, 0.1) 1px, transparent 1px),
    linear-gradient(rgba(200, 200, 200, 0.1) 1px, transparent 1px);
  background-size: 20px 20px;
  padding: 20px;
  overflow-y: auto;
}

:deep(.el-menu-item) {
  border-radius: 6px;
  margin: 4px 8px;
}

:deep(.el-menu-item.is-active) {
  background-color: #4299E1 !important;
  color: #FFFFFF !important;
}

:deep(.el-breadcrumb__inner) {
  color: #4A5568;
}

:deep(.el-breadcrumb__inner.is-link) {
  color: #4299E1;
}
</style>
