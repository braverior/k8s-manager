# K8s 管理工具 - API 接口文档

## 基础地址

```
http://localhost:8080/api/v1
```

## 响应格式

所有响应遵循统一格式：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

分页响应格式：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 100,
    "items": []
  }
}
```

## 错误码

| 错误码 | HTTP 状态码 | 描述 |
|--------|-------------|------|
| 0 | 200 | 成功 |
| 400 | 400 | 请求参数错误 |
| 401 | 401 | 未授权（未登录或 Token 失效） |
| 403 | 403 | 无权限访问 |
| 404 | 404 | 资源不存在 |
| 409 | 409 | 资源冲突 |
| 500 | 500 | 服务器内部错误 |
| 4001 | 404 | 集群不存在 |
| 4002 | 400 | 集群连接失败 |
| 4003 | 404 | Kubernetes 资源不存在 |
| 4004 | 404 | 历史记录不存在 |
| 4005 | 403 | 无权限访问该集群 |
| 4006 | 403 | 无权限访问该命名空间 |
| 4010 | 401 | 飞书授权码无效或已过期 |
| 4011 | 401 | 获取飞书 Token 失败 |
| 4012 | 401 | 获取飞书用户信息失败 |
| 4013 | 403 | 用户无权限访问本系统 |
| 4014 | 403 | 用户已被禁用 |
| 4020 | 404 | 用户不存在 |
| 4021 | 403 | 需要管理员权限 |
| 4022 | 400 | 不能修改自己的管理员角色 |

---

## 集群管理

> **注意**：集群配置从配置文件加载，不支持通过 API 创建、更新或删除集群。

### 获取集群列表

```
GET /clusters
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "production",
      "description": "生产环境集群",
      "status": "connected"
    },
    {
      "name": "staging",
      "description": "预发布环境集群",
      "status": "connected"
    }
  ]
}
```

### 获取集群详情

```
GET /clusters/:cluster
```

**路径参数：**
| 参数 | 类型 | 描述 |
|------|------|------|
| cluster | string | 集群名称 |

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "name": "production",
    "description": "生产环境集群",
    "status": "connected"
  }
}
```

### 测试集群连接

```
POST /clusters/:cluster/test-connection
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "success": true,
    "message": "连接成功",
    "version": "v1.28.0"
  }
}
```

### 获取命名空间列表

```
GET /clusters/:cluster/namespaces
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "default",
      "status": "Active"
    },
    {
      "name": "kube-system",
      "status": "Active"
    }
  ]
}
```

### 获取集群 Dashboard（概览信息）

```
GET /clusters/:cluster/dashboard
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "cluster_name": "production",
    "status": "connected",
    "version": "v1.28.0",
    "resources": {
      "nodes": 3,
      "namespaces": 10,
      "pods": {
        "total": 50,
        "running": 45,
        "pending": 2,
        "succeeded": 3,
        "failed": 0
      },
      "deployments": 15,
      "services": 20,
      "configmaps": 30,
      "hpas": 5
    },
    "capacity": {
      "cpu": "4",
      "memory": "16Gi",
      "pods": 110
    },
    "usage": {
      "cpu": "500m",
      "cpu_percentage": 12.5,
      "memory": "4Gi",
      "memory_percentage": 25.0
    },
    "hpa_summaries": [
      {
        "name": "nginx-hpa",
        "namespace": "default",
        "target_kind": "Deployment",
        "target_name": "nginx",
        "min_replicas": 1,
        "max_replicas": 10,
        "current_replicas": 3,
        "desired_replicas": 3,
        "cpu_target_utilization": 80,
        "cpu_current_utilization": 45,
        "memory_target_utilization": null,
        "memory_current_utilization": null
      }
    ]
  }
}
```

**字段说明：**
| 字段 | 类型 | 描述 |
|------|------|------|
| cluster_name | string | 集群名称 |
| status | string | 集群连接状态（connected/disconnected） |
| version | string | Kubernetes 版本 |
| resources | object | 资源数量统计 |
| resources.nodes | int | 节点数量 |
| resources.namespaces | int | 命名空间数量 |
| resources.pods | object | Pod 状态统计 |
| resources.pods.total | int | Pod 总数 |
| resources.pods.running | int | 运行中的 Pod 数量 |
| resources.pods.pending | int | 等待中的 Pod 数量 |
| resources.pods.succeeded | int | 成功完成的 Pod 数量 |
| resources.pods.failed | int | 失败的 Pod 数量 |
| resources.deployments | int | Deployment 数量 |
| resources.services | int | Service 数量 |
| resources.configmaps | int | ConfigMap 数量 |
| resources.hpas | int | HPA 数量 |
| capacity | object | 集群资源总容量 |
| capacity.cpu | string | CPU 总容量 |
| capacity.memory | string | 内存总容量 |
| capacity.pods | int | 可调度 Pod 总数 |
| usage | object | 集群资源使用量（需要 metrics-server） |
| usage.cpu | string | CPU 已使用量 |
| usage.cpu_percentage | float | CPU 使用百分比 |
| usage.memory | string | 内存已使用量 |
| usage.memory_percentage | float | 内存使用百分比 |
| hpa_summaries | array | HPA 概要列表 |
| hpa_summaries[].name | string | HPA 名称 |
| hpa_summaries[].namespace | string | 命名空间 |
| hpa_summaries[].target_kind | string | 目标资源类型（如 Deployment） |
| hpa_summaries[].target_name | string | 目标资源名称 |
| hpa_summaries[].min_replicas | int | 最小副本数 |
| hpa_summaries[].max_replicas | int | 最大副本数 |
| hpa_summaries[].current_replicas | int | 当前副本数 |
| hpa_summaries[].desired_replicas | int | 期望副本数 |
| hpa_summaries[].cpu_target_utilization | int | CPU 目标利用率（百分比） |
| hpa_summaries[].cpu_current_utilization | int | CPU 当前利用率（百分比） |
| hpa_summaries[].memory_target_utilization | int | 内存目标利用率（百分比） |
| hpa_summaries[].memory_current_utilization | int | 内存当前利用率（百分比） |

> **注意**：`usage` 字段需要集群安装 metrics-server，若未安装则该字段为 null。`capacity` 字段始终返回。`hpa_summaries` 中的利用率字段可能为 null（如果 HPA 未配置对应的指标）。

---

## 节点管理

> 节点是集群级别的只读资源，不支持通过 API 创建或删除。

### 获取节点列表

```
GET /clusters/:cluster/nodes
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "node-1",
      "status": "Ready",
      "roles": ["control-plane", "master"],
      "internal_ip": "192.168.1.100",
      "kubelet_version": "v1.28.0",
      "cpu_capacity": "4",
      "memory_capacity": "16Gi",
      "cpu_usage": "800m",
      "cpu_percentage": 20.0,
      "memory_usage": "4Gi",
      "memory_percentage": 25.0,
      "pod_count": 25,
      "created_at": "2024-01-01T00:00:00Z"
    },
    {
      "name": "node-2",
      "status": "Ready",
      "roles": ["worker"],
      "internal_ip": "192.168.1.101",
      "kubelet_version": "v1.28.0",
      "cpu_capacity": "4",
      "memory_capacity": "16Gi",
      "pod_count": 20,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**字段说明：**
| 字段 | 类型 | 描述 |
|------|------|------|
| name | string | 节点名称 |
| status | string | 节点状态（Ready/NotReady/Unknown） |
| roles | array | 节点角色列表 |
| internal_ip | string | 内部 IP 地址 |
| kubelet_version | string | Kubelet 版本 |
| cpu_capacity | string | CPU 可分配容量 |
| memory_capacity | string | 内存可分配容量 |
| cpu_usage | string | CPU 使用量（需要 metrics-server） |
| cpu_percentage | float | CPU 使用百分比 |
| memory_usage | string | 内存使用量（需要 metrics-server） |
| memory_percentage | float | 内存使用百分比 |
| pod_count | int | 运行的 Pod 数量 |
| created_at | string | 节点创建时间 |

### 获取节点详情

```
GET /clusters/:cluster/nodes/:name
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "name": "node-1",
    "status": "Ready",
    "roles": ["control-plane", "master"],
    "internal_ip": "192.168.1.100",
    "external_ip": "",
    "kubelet_version": "v1.28.0",
    "kernel_version": "5.15.0-91-generic",
    "os_image": "Ubuntu 22.04.3 LTS",
    "os": "linux",
    "architecture": "amd64",
    "container_runtime": "containerd://1.7.0",
    "created_at": "2024-01-01T00:00:00Z",
    "capacity": {
      "cpu": "4",
      "memory": "16Gi",
      "pods": "110",
      "ephemeral_storage": "100Gi"
    },
    "allocatable": {
      "cpu": "3800m",
      "memory": "15Gi",
      "pods": "110",
      "ephemeral_storage": "90Gi"
    },
    "usage": {
      "cpu": "800m",
      "cpu_percentage": 21.05,
      "memory": "4Gi",
      "memory_percentage": 26.67
    },
    "conditions": [
      {
        "type": "Ready",
        "status": "True",
        "reason": "KubeletReady",
        "message": "kubelet is posting ready status",
        "last_heartbeat_time": "2024-01-15T10:30:00Z"
      },
      {
        "type": "MemoryPressure",
        "status": "False",
        "reason": "KubeletHasSufficientMemory",
        "message": "kubelet has sufficient memory available"
      }
    ],
    "taints": [
      {
        "key": "node-role.kubernetes.io/control-plane",
        "effect": "NoSchedule"
      }
    ],
    "labels": {
      "kubernetes.io/arch": "amd64",
      "kubernetes.io/os": "linux",
      "node-role.kubernetes.io/control-plane": ""
    },
    "pod_count": 25
  }
}
```

**字段说明：**
| 字段 | 类型 | 描述 |
|------|------|------|
| capacity | object | 节点资源总容量 |
| allocatable | object | 节点可分配资源（扣除系统预留） |
| usage | object | 节点资源使用情况（需要 metrics-server） |
| conditions | array | 节点条件状态列表 |
| taints | array | 节点污点列表 |
| labels | object | 节点标签 |
| pod_count | int | 该节点上运行的 Pod 数量 |

---

## 资源管理通用说明

### 请求格式

所有资源（ConfigMap、Deployment、Service、HPA）的创建和更新操作使用统一的请求格式：

```json
{
  "yaml": "完整的 YAML 内容",
  "content": "Base64 编码的 YAML 内容"
}
```

**字段说明：**
- `yaml`：直接提供 YAML 格式的资源定义
- `content`：Base64 编码后的 YAML 内容

> **注意**：`yaml` 和 `content` 二选一即可，优先使用 `yaml` 字段。

### 响应格式

所有资源的查询和操作响应统一返回干净的 YAML 格式（已清理系统自动添加的字段）：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "name": "资源名称",
    "namespace": "命名空间",
    "yaml": "完整的 YAML 内容"
  }
}
```

**已清理的系统字段：**
- `metadata.creationTimestamp` - 创建时间戳
- `metadata.resourceVersion` - 资源版本
- `metadata.uid` - 唯一标识
- `metadata.generation` - 代数
- `metadata.managedFields` - 托管字段
- `metadata.selfLink` - 自链接
- `metadata.finalizers` - 终结器
- `metadata.ownerReferences` - 所有者引用
- `status` - 状态信息

> **说明**：返回的 YAML 可以直接用于修改后重新提交，无需手动删除系统字段。

列表查询返回数组格式：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "资源名称",
      "namespace": "命名空间",
      "yaml": "完整的 YAML 内容"
    }
  ]
}
```

### YAML 示例

**ConfigMap：**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  key1: value1
  key2: value2
```

**Deployment：**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
        ports:
        - containerPort: 80
```

**Service：**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
spec:
  type: ClusterIP
  selector:
    app: nginx
  ports:
  - port: 80
    targetPort: 80
```

**HPA：**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: nginx-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: nginx
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 80
```

---

## ConfigMap 管理

### 获取 ConfigMap 列表

```
GET /clusters/:cluster/namespaces/:namespace/configmaps
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "app-config",
      "namespace": "default",
      "yaml": "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: app-config\n  namespace: default\ndata:\n  key1: value1\n  key2: value2\n"
    }
  ]
}
```

### 创建/更新 ConfigMap (Apply)

```
POST /clusters/:cluster/namespaces/:namespace/configmaps
PUT /clusters/:cluster/namespaces/:namespace/configmaps/:name
```

**请求参数：**
```json
{
  "yaml": "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: app-config\ndata:\n  key1: value1\n  key2: value2"
}
```

或使用 Base64 编码：
```json
{
  "content": "YXBpVmVyc2lvbjogdjEKa2luZDogQ29uZmlnTWFwCm1ldGFkYXRhOgogIG5hbWU6IGFwcC1jb25maWcKZGF0YToKICBrZXkxOiB2YWx1ZTEKICBrZXkyOiB2YWx1ZTI="
}
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "name": "app-config",
    "namespace": "default",
    "yaml": "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: app-config\n  namespace: default\ndata:\n  key1: value1\n  key2: value2\n"
  }
}
```

### 获取 ConfigMap 详情

```
GET /clusters/:cluster/namespaces/:namespace/configmaps/:name
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "name": "app-config",
    "namespace": "default",
    "yaml": "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: app-config\n  namespace: default\ndata:\n  key1: value1\n  key2: value2\n"
  }
}
```

### 删除 ConfigMap

```
DELETE /clusters/:cluster/namespaces/:namespace/configmaps/:name
```

---

## Deployment 管理

### 获取 Deployment 列表

```
GET /clusters/:cluster/namespaces/:namespace/deployments
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "nginx",
      "namespace": "default",
      "replicas": 3,
      "ready_replicas": 3,
      "available_replicas": 3,
      "pod_count": 3,
      "yaml": "apiVersion: apps/v1\nkind: Deployment\n..."
    }
  ]
}
```

**字段说明：**
| 字段 | 类型 | 描述 |
|------|------|------|
| name | string | Deployment 名称 |
| namespace | string | 命名空间 |
| replicas | int | 期望副本数 |
| ready_replicas | int | 就绪副本数 |
| available_replicas | int | 可用副本数 |
| pod_count | int | 实际 Pod 数量 |
| yaml | string | 完整 YAML 内容 |

### 创建/更新 Deployment (Apply)

```
POST /clusters/:cluster/namespaces/:namespace/deployments
PUT /clusters/:cluster/namespaces/:namespace/deployments/:name
```

**请求参数：**
```json
{
  "yaml": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: nginx\n..."
}
```

### 获取 Deployment 详情

```
GET /clusters/:cluster/namespaces/:namespace/deployments/:name
```

### 删除 Deployment

```
DELETE /clusters/:cluster/namespaces/:namespace/deployments/:name
```

### 获取 Deployment 关联的 Pod

```
GET /clusters/:cluster/namespaces/:namespace/deployments/:name/pods
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "nginx-7c5b4bf8b9-abc12",
      "namespace": "default",
      "phase": "Running",
      "pod_ip": "10.244.1.10",
      "host_ip": "192.168.1.100",
      "node_name": "node-1",
      "ready_containers": 1,
      "total_containers": 1,
      "restart_count": 0,
      "created_at": "2024-01-15T10:30:00Z",
      "containers": [
        {
          "name": "nginx",
          "ready": true,
          "restart_count": 0,
          "state": "Running",
          "image": "nginx:1.21",
          "started_at": "2024-01-15T10:30:05Z"
        }
      ],
      "metrics": {
        "containers": [
          {
            "name": "nginx",
            "cpu": "10m",
            "memory": "64Mi"
          }
        ]
      }
    }
  ]
}
```

---

## Pod 管理

> Pod 是只读资源，不支持直接创建或更新。可以通过删除 Pod 来触发 Deployment 重新调度。

### 获取 Pod 列表

```
GET /clusters/:cluster/namespaces/:namespace/pods
```

**查询参数：**
| 参数 | 类型 | 是否必填 | 描述 |
|------|------|----------|------|
| deployment | string | 否 | 按 Deployment 名称筛选 Pod |

**示例：**
```
GET /clusters/local/namespaces/default/pods?deployment=nginx
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "nginx-7c5b4bf8b9-abc12",
      "namespace": "default",
      "phase": "Running",
      "pod_ip": "10.244.1.10",
      "host_ip": "192.168.1.100",
      "node_name": "node-1",
      "ready_containers": 1,
      "total_containers": 1,
      "restart_count": 0,
      "created_at": "2024-01-15T10:30:00Z",
      "containers": [...],
      "metrics": {...}
    }
  ]
}
```

### 获取 Pod 详情

```
GET /clusters/:cluster/namespaces/:namespace/pods/:name
```

### 删除 Pod（重启）

```
DELETE /clusters/:cluster/namespaces/:namespace/pods/:name
```

> **说明**：删除由 Deployment/ReplicaSet 管理的 Pod 会自动创建新的 Pod，达到重启效果。

---

## Service 管理

### 获取 Service 列表

```
GET /clusters/:cluster/namespaces/:namespace/services
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "nginx-service",
      "namespace": "default",
      "yaml": "apiVersion: v1\nkind: Service\nmetadata:\n  name: nginx-service\n  namespace: default\nspec:\n  type: ClusterIP\n  ..."
    }
  ]
}
```

### 创建/更新 Service (Apply)

```
POST /clusters/:cluster/namespaces/:namespace/services
PUT /clusters/:cluster/namespaces/:namespace/services/:name
```

**请求参数：**
```json
{
  "yaml": "apiVersion: v1\nkind: Service\nmetadata:\n  name: nginx-service\n..."
}
```

### 获取 Service 详情

```
GET /clusters/:cluster/namespaces/:namespace/services/:name
```

### 删除 Service

```
DELETE /clusters/:cluster/namespaces/:namespace/services/:name
```

---

## HPA 管理

### 获取 HPA 列表

```
GET /clusters/:cluster/namespaces/:namespace/hpas
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "nginx-hpa",
      "namespace": "default",
      "scale_target_ref": {
        "api_version": "apps/v1",
        "kind": "Deployment",
        "name": "nginx"
      },
      "min_replicas": 1,
      "max_replicas": 10,
      "current_replicas": 3,
      "desired_replicas": 3,
      "cpu": {
        "target_type": "Utilization",
        "target_utilization": 80,
        "current_utilization": 45,
        "current_average_value": "150m"
      },
      "memory": {
        "target_type": "Utilization",
        "target_utilization": 70,
        "current_utilization": 55,
        "current_average_value": "256Mi"
      },
      "metrics": [
        {
          "type": "Resource",
          "name": "cpu",
          "target_type": "Utilization",
          "target_utilization": 80,
          "current_utilization": 45
        },
        {
          "type": "Resource",
          "name": "memory",
          "target_type": "Utilization",
          "target_utilization": 70,
          "current_utilization": 55
        }
      ],
      "conditions": [
        {
          "type": "AbleToScale",
          "status": "True",
          "reason": "ReadyForNewScale",
          "message": "recommended size matches current size",
          "last_transition_time": "2024-01-15T10:30:00Z"
        },
        {
          "type": "ScalingActive",
          "status": "True",
          "reason": "ValidMetricFound",
          "message": "the HPA was able to successfully calculate a replica count",
          "last_transition_time": "2024-01-15T10:30:00Z"
        }
      ],
      "yaml": "apiVersion: autoscaling/v2\nkind: HorizontalPodAutoscaler\n..."
    }
  ]
}
```

**字段说明：**
| 字段 | 类型 | 描述 |
|------|------|------|
| name | string | HPA 名称 |
| namespace | string | 命名空间 |
| scale_target_ref | object | 目标资源引用 |
| scale_target_ref.api_version | string | 目标资源 API 版本 |
| scale_target_ref.kind | string | 目标资源类型（如 Deployment） |
| scale_target_ref.name | string | 目标资源名称 |
| min_replicas | int | 最小副本数 |
| max_replicas | int | 最大副本数 |
| current_replicas | int | 当前副本数 |
| desired_replicas | int | 期望副本数 |
| cpu | object | CPU 资源指标（如果配置了 CPU 指标） |
| cpu.target_type | string | 目标类型（Utilization/AverageValue） |
| cpu.target_utilization | int | CPU 目标利用率百分比 |
| cpu.target_average_value | string | CPU 目标平均值（如 "500m"） |
| cpu.current_utilization | int | CPU 当前利用率百分比 |
| cpu.current_average_value | string | CPU 当前平均值 |
| memory | object | 内存资源指标（如果配置了内存指标） |
| memory.target_type | string | 目标类型（Utilization/AverageValue） |
| memory.target_utilization | int | 内存目标利用率百分比 |
| memory.target_average_value | string | 内存目标平均值（如 "200Mi"） |
| memory.current_utilization | int | 内存当前利用率百分比 |
| memory.current_average_value | string | 内存当前平均值 |
| metrics | array | 所有指标配置和当前值列表 |
| metrics[].type | string | 指标类型（Resource/Pods/Object/External） |
| metrics[].name | string | 指标名称（如 cpu, memory） |
| metrics[].target_type | string | 目标类型（Utilization/AverageValue/Value） |
| metrics[].target_value | string | 目标值（非利用率类型时） |
| metrics[].target_utilization | int | 目标利用率百分比 |
| metrics[].current_value | string | 当前值（非利用率类型时） |
| metrics[].current_utilization | int | 当前利用率百分比 |
| conditions | array | HPA 条件状态列表 |
| conditions[].type | string | 条件类型 |
| conditions[].status | string | 状态（True/False/Unknown） |
| conditions[].reason | string | 原因 |
| conditions[].message | string | 详细消息 |
| conditions[].last_transition_time | string | 最后转换时间 |
| yaml | string | 完整 YAML 内容 |

> **注意**：`cpu` 和 `memory` 字段仅在 HPA 配置了对应的资源指标时才会返回。如果 HPA 只配置了 CPU 指标，则 `memory` 字段为 null。

### 获取 HPA 详情

```
GET /clusters/:cluster/namespaces/:namespace/hpas/:name
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "name": "nginx-hpa",
    "namespace": "default",
    "scale_target_ref": {
      "api_version": "apps/v1",
      "kind": "Deployment",
      "name": "nginx"
    },
    "min_replicas": 1,
    "max_replicas": 10,
    "current_replicas": 3,
    "desired_replicas": 3,
    "cpu": {
      "target_type": "Utilization",
      "target_utilization": 80,
      "current_utilization": 45,
      "current_average_value": "150m"
    },
    "memory": {
      "target_type": "Utilization",
      "target_utilization": 70,
      "current_utilization": 55,
      "current_average_value": "256Mi"
    },
    "metrics": [...],
    "conditions": [...],
    "yaml": "apiVersion: autoscaling/v2\nkind: HorizontalPodAutoscaler\n..."
  }
}
```

### 创建/更新 HPA (Apply)

```
POST /clusters/:cluster/namespaces/:namespace/hpas
PUT /clusters/:cluster/namespaces/:namespace/hpas/:name
```

**请求参数：**
```json
{
  "yaml": "apiVersion: autoscaling/v2\nkind: HorizontalPodAutoscaler\nmetadata:\n  name: nginx-hpa\n..."
}
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "name": "nginx-hpa",
    "namespace": "default",
    "yaml": "apiVersion: autoscaling/v2\nkind: HorizontalPodAutoscaler\n..."
  }
}
```

### 删除 HPA

```
DELETE /clusters/:cluster/namespaces/:namespace/hpas/:name
```

---

## 历史版本管理

> 系统会自动保存所有资源的变更历史，支持查询、对比和回滚操作。

### 获取历史记录列表

```
GET /clusters/:cluster/namespaces/:namespace/histories
```

**查询参数：**
| 参数 | 类型 | 描述 |
|------|------|------|
| resource_type | string | 按资源类型过滤（ConfigMap, Deployment, Service, HPA） |
| resource_name | string | 按资源名称过滤 |
| page | int | 页码（默认：1） |
| page_size | int | 每页数量（默认：20） |

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 50,
    "items": [
      {
        "id": 1,
        "cluster_name": "production",
        "namespace": "default",
        "resource_type": "ConfigMap",
        "resource_name": "app-config",
        "version": 3,
        "operation": "update",
        "operator": "admin",
        "created_at": "2024-01-15T10:30:00Z"
      }
    ]
  }
}
```

### 获取历史记录详情

```
GET /clusters/:cluster/namespaces/:namespace/histories/:id
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "cluster_name": "production",
    "namespace": "default",
    "resource_type": "ConfigMap",
    "resource_name": "app-config",
    "version": 3,
    "operation": "update",
    "operator": "admin",
    "created_at": "2024-01-15T10:30:00Z",
    "content": "{...}"
  }
}
```

### 版本对比（Diff）

```
GET /clusters/:cluster/namespaces/:namespace/histories/diff
```

**查询参数：**
| 参数 | 类型 | 是否必填 | 描述 |
|------|------|----------|------|
| source_version | uint64 | 是 | 源版本历史记录 ID |
| target_version | uint64 | 是 | 目标版本历史记录 ID |

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "source_version": 1,
    "target_version": 2,
    "source_content": "{...}",
    "target_content": "{...}",
    "diff": "--- source\n+++ target\n- 旧内容\n+ 新内容"
  }
}
```

### 回滚到指定版本

```
POST /clusters/:cluster/namespaces/:namespace/histories/:id/rollback
```

**请求参数（可选）：**
```json
{
  "operator": "admin"
}
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "success": true,
    "message": "成功回滚到版本 2",
    "restored_version": 2,
    "new_version": 4
  }
}
```

---

## 认证

### 飞书登录流程

```
┌────────┐      ┌────────┐      ┌────────┐      ┌────────┐
│ 前端 A │      │ 飞书   │      │ 后端 B │      │ 飞书API │
└───┬────┘      └───┬────┘      └───┬────┘      └───┬────┘
    │               │               │               │
    │ 1. 获取登录配置│               │               │
    │──────────────────────────────>│               │
    │<──────────────────────────────│               │
    │ (返回 app_id, redirect_uri)   │               │
    │               │               │               │
    │ 2. 跳转飞书授权页面            │               │
    │──────────────>│               │               │
    │               │               │               │
    │ 3. 用户授权后回调(带 code)     │               │
    │<──────────────│               │               │
    │               │               │               │
    │ 4. 用 code 换取 token         │               │
    │──────────────────────────────>│               │
    │               │               │ 5. 换取飞书 token
    │               │               │──────────────>│
    │               │               │<──────────────│
    │               │               │ 6. 获取用户信息
    │               │               │──────────────>│
    │               │               │<──────────────│
    │<──────────────────────────────│               │
    │ (返回系统 token + 用户信息)    │               │
    └               └               └               └
```

### 获取飞书登录配置

前端调用此接口获取构建飞书授权 URL 所需的参数。

```
GET /auth/feishu/config
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "app_id": "cli_xxxxxxxxxx",
    "redirect_uri": "https://domain-a.com/auth/feishu/callback",
    "authorize_url": "https://passport.feishu.cn/suite/passport/oauth/authorize"
  }
}
```

**字段说明：**
| 字段 | 类型 | 描述 |
|------|------|------|
| app_id | string | 飞书应用的 App ID |
| redirect_uri | string | 授权回调地址（前端域名 A） |
| authorize_url | string | 飞书授权页面基础 URL |

**前端构建授权 URL：**
```
{authorize_url}?client_id={app_id}&redirect_uri={redirect_uri}&response_type=code&state={随机字符串}
```

> **注意**：`state` 参数用于防止 CSRF 攻击，前端应生成随机字符串并在回调时验证。

---

### 飞书登录（Code 换 Token）

前端获取到飞书回调的 authorization code 后，调用此接口完成登录。

```
POST /auth/feishu/login
```

**请求参数：**
```json
{
  "code": "飞书返回的授权码",
  "state": "前端生成的随机字符串（可选，用于安全校验）"
}
```

**字段说明：**
| 字段 | 类型 | 是否必填 | 描述 |
|------|------|----------|------|
| code | string | 是 | 飞书授权回调返回的 authorization code |
| state | string | 否 | 前端生成的 state 参数（建议传入用于校验） |

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400,
    "user": {
      "id": "ou_xxxxxxxxxx",
      "name": "张三",
      "email": "zhangsan@example.com",
      "avatar_url": "https://internal-api-lark-file.feishu.cn/avatar/xxx",
      "mobile": "13800138000",
      "employee_id": "EMP001",
      "department": {
        "id": "od_xxxxxxxxxx",
        "name": "技术部",
        "path": "公司/研发中心/技术部"
      },
      "role": "user",
      "is_admin": false,
      "permissions": [
        {
          "cluster": "production",
          "namespaces": ["default", "app-prod"]
        },
        {
          "cluster": "staging",
          "namespaces": ["*"]
        }
      ]
    }
  }
}
```

**字段说明：**
| 字段 | 类型 | 描述 |
|------|------|------|
| token | string | 系统生成的 JWT Token，用于后续 API 调用 |
| expires_in | int | Token 有效期（秒） |
| user.id | string | 飞书用户唯一标识（open_id 或 user_id） |
| user.name | string | 用户姓名 |
| user.email | string | 用户邮箱（可能为空） |
| user.avatar_url | string | 用户头像 URL |
| user.mobile | string | 用户手机号（可能为空，取决于授权范围） |
| user.employee_id | string | 企业内员工 ID（可能为空） |
| user.department | object | 用户所属部门信息 |
| user.department.id | string | 部门 ID |
| user.department.name | string | 部门名称 |
| user.department.path | string | 部门完整路径 |
| user.role | string | 用户角色（admin/user） |
| user.is_admin | bool | 是否为管理员 |
| user.permissions | array | 用户权限列表（管理员拥有所有权限） |
| user.permissions[].cluster | string | 可访问的集群名称 |
| user.permissions[].namespaces | array | 可访问的命名空间列表，`["*"]` 表示所有命名空间 |

**错误响应：**

| 错误码 | HTTP 状态码 | 描述 |
|--------|-------------|------|
| 4010 | 401 | 授权码无效或已过期 |
| 4011 | 401 | 获取飞书 Token 失败 |
| 4012 | 401 | 获取飞书用户信息失败 |
| 4013 | 403 | 用户无权限访问本系统 |

```json
{
  "code": 4010,
  "message": "授权码无效或已过期",
  "data": null
}
```

---

### 获取当前用户信息

已登录用户获取自己的信息和权限。

```
GET /auth/me
```

**请求头：**
```
Authorization: Bearer {token}
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "ou_xxxxxxxxxx",
    "name": "张三",
    "email": "zhangsan@example.com",
    "avatar_url": "https://internal-api-lark-file.feishu.cn/avatar/xxx",
    "mobile": "13800138000",
    "employee_id": "EMP001",
    "department": {
      "id": "od_xxxxxxxxxx",
      "name": "技术部",
      "path": "公司/研发中心/技术部"
    },
    "role": "user",
    "is_admin": false,
    "permissions": [
      {
        "cluster": "production",
        "namespaces": ["default", "app-prod"]
      },
      {
        "cluster": "staging",
        "namespaces": ["*"]
      }
    ]
  }
}
```

---

### 退出登录

```
POST /auth/logout
```

**请求头：**
```
Authorization: Bearer {token}
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 用户管理（管理员）

> 以下接口仅管理员可访问，用于管理系统用户和权限。

### 获取用户列表

```
GET /admin/users
```

**请求头：**
```
Authorization: Bearer {token}
```

**查询参数：**
| 参数 | 类型 | 是否必填 | 描述 |
|------|------|----------|------|
| keyword | string | 否 | 搜索关键词（匹配姓名、邮箱、员工 ID） |
| department_id | string | 否 | 按部门 ID 筛选 |
| role | string | 否 | 按角色筛选（admin/user） |
| page | int | 否 | 页码（默认：1） |
| page_size | int | 否 | 每页数量（默认：20） |

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 50,
    "items": [
      {
        "id": "ou_xxxxxxxxxx",
        "name": "张三",
        "email": "zhangsan@example.com",
        "avatar_url": "https://internal-api-lark-file.feishu.cn/avatar/xxx",
        "mobile": "13800138000",
        "employee_id": "EMP001",
        "department": {
          "id": "od_xxxxxxxxxx",
          "name": "技术部",
          "path": "公司/研发中心/技术部"
        },
        "role": "user",
        "is_admin": false,
        "status": "active",
        "last_login_at": "2024-01-15T10:30:00Z",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

**字段说明：**
| 字段 | 类型 | 描述 |
|------|------|------|
| status | string | 用户状态（active-正常/disabled-禁用） |
| last_login_at | string | 最后登录时间 |
| created_at | string | 首次登录时间（用户首次登录后自动创建） |

---

### 获取用户详情

```
GET /admin/users/:user_id
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "ou_xxxxxxxxxx",
    "name": "张三",
    "email": "zhangsan@example.com",
    "avatar_url": "https://internal-api-lark-file.feishu.cn/avatar/xxx",
    "mobile": "13800138000",
    "employee_id": "EMP001",
    "department": {
      "id": "od_xxxxxxxxxx",
      "name": "技术部",
      "path": "公司/研发中心/技术部"
    },
    "role": "user",
    "is_admin": false,
    "status": "active",
    "permissions": [
      {
        "cluster": "production",
        "namespaces": ["default", "app-prod"]
      },
      {
        "cluster": "staging",
        "namespaces": ["*"]
      }
    ],
    "last_login_at": "2024-01-15T10:30:00Z",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### 更新用户角色

设置用户为管理员或普通用户。

```
PUT /admin/users/:user_id/role
```

**请求参数：**
```json
{
  "role": "admin"
}
```

**字段说明：**
| 字段 | 类型 | 是否必填 | 描述 |
|------|------|----------|------|
| role | string | 是 | 用户角色（admin/user） |

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

> **注意**：管理员拥有所有集群和命名空间的访问权限，无需单独配置 permissions。

---

### 启用/禁用用户

```
PUT /admin/users/:user_id/status
```

**请求参数：**
```json
{
  "status": "disabled"
}
```

**字段说明：**
| 字段 | 类型 | 是否必填 | 描述 |
|------|------|----------|------|
| status | string | 是 | 用户状态（active/disabled） |

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

> **注意**：禁用用户后，该用户将无法登录系统，已登录的会话也会失效。

---

## 权限管理（管理员）

> 管理用户对集群和命名空间的访问权限。

### 获取用户权限

```
GET /admin/users/:user_id/permissions
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "ou_xxxxxxxxxx",
    "user_name": "张三",
    "is_admin": false,
    "permissions": [
      {
        "cluster": "production",
        "namespaces": ["default", "app-prod"]
      },
      {
        "cluster": "staging",
        "namespaces": ["*"]
      }
    ]
  }
}
```

---

### 设置用户权限

覆盖用户的所有权限配置。

```
PUT /admin/users/:user_id/permissions
```

**请求参数：**
```json
{
  "permissions": [
    {
      "cluster": "production",
      "namespaces": ["default", "app-prod", "app-test"]
    },
    {
      "cluster": "staging",
      "namespaces": ["*"]
    }
  ]
}
```

**字段说明：**
| 字段 | 类型 | 是否必填 | 描述 |
|------|------|----------|------|
| permissions | array | 是 | 权限配置列表 |
| permissions[].cluster | string | 是 | 集群名称 |
| permissions[].namespaces | array | 是 | 允许访问的命名空间列表，`["*"]` 表示该集群所有命名空间 |

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 添加集群权限

为用户添加某个集群的访问权限。

```
POST /admin/users/:user_id/permissions/clusters
```

**请求参数：**
```json
{
  "cluster": "production",
  "namespaces": ["default", "app-prod"]
}
```

**字段说明：**
| 字段 | 类型 | 是否必填 | 描述 |
|------|------|----------|------|
| cluster | string | 是 | 集群名称 |
| namespaces | array | 是 | 允许访问的命名空间列表 |

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 移除集群权限

移除用户对某个集群的访问权限。

```
DELETE /admin/users/:user_id/permissions/clusters/:cluster
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 更新集群命名空间权限

更新用户在某个集群下可访问的命名空间。

```
PUT /admin/users/:user_id/permissions/clusters/:cluster/namespaces
```

**请求参数：**
```json
{
  "namespaces": ["default", "app-prod", "app-test"]
}
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 批量设置用户权限

为多个用户批量设置相同的权限（适用于团队配置）。

```
POST /admin/permissions/batch
```

**请求参数：**
```json
{
  "user_ids": ["ou_xxx1", "ou_xxx2", "ou_xxx3"],
  "permissions": [
    {
      "cluster": "staging",
      "namespaces": ["*"]
    }
  ]
}
```

**响应示例：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "success_count": 3,
    "failed_count": 0,
    "failed_users": []
  }
}
```

---

## 健康检查

```
GET /health
```

**响应示例：**
```json
{
  "status": "ok"
}
```
