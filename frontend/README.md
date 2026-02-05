# K8S Manager Frontend

一个现代化的 Kubernetes 管理平台前端，使用 React + TypeScript + Tailwind CSS + shadcn/ui 构建。

## 技术栈

- **React 18** - UI框架
- **TypeScript** - 类型安全
- **Tailwind CSS** - 样式框架
- **shadcn/ui** - UI组件库
- **Monaco Editor** - YAML编辑器
- **React Router** - 路由管理
- **Vite** - 构建工具

## 功能特性

- **Dashboard** - 集群概览和快速导航
- **ConfigMaps 管理** - 创建、查看、编辑、删除 ConfigMap
- **Deployments 管理** - 创建、查看、编辑、删除 Deployment
- **History 历史记录** - 查看资源变更历史

## 快速开始

### 安装依赖

```bash
cd frontend
npm install
```

### 开发模式

```bash
npm run dev
```

访问 http://localhost:3000

### 生产构建

```bash
npm run build
```

构建产物在 `dist` 目录

## 项目结构

```
frontend/
├── src/
│   ├── api/              # API 请求封装
│   ├── components/       # 公共组件
│   │   ├── ui/          # shadcn/ui 组件
│   │   ├── Layout.tsx   # 布局组件
│   │   ├── Sidebar.tsx  # 侧边栏
│   │   └── ...
│   ├── hooks/           # 自定义 Hooks
│   ├── lib/             # 工具函数
│   ├── pages/           # 页面组件
│   ├── types/           # TypeScript 类型定义
│   ├── App.tsx          # 根组件
│   ├── main.tsx         # 入口文件
│   └── index.css        # 全局样式
├── public/              # 静态资源
├── index.html           # HTML 模板
└── package.json
```

## API 配置

默认 API 地址为 `http://127.0.0.1:30001/api/v1`，可在 `vite.config.ts` 中修改代理配置。

## 设计风格

- 深色主题（Dark Mode OLED）
- 专业 DevOps 风格
- Fira Code / Fira Sans 字体
- 绿色主题色（#22C55E）
