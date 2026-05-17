# Docker 磁盘空间管理

## ⚠️ 问题说明

Docker 在构建镜像时会生成大量 **Build Cache**（构建缓存），这些缓存默认存储在系统盘（C:\），会随着时间不断累积，最终可能导致系统盘空间耗尽。

## 📁 清理脚本

### 1. **build-and-cleanup.bat** - 推荐使用
智能构建脚本，在构建后自动清理缓存。

**使用方法：**
```bash
# 在项目根目录运行
build-and-cleanup.bat
```

**功能：**
- ✅ 构建前清理缓存
- ✅ 执行 docker-compose build
- ✅ 自动启动容器
- ✅ 构建后自动清理缓存
- ✅ 显示清理前后的磁盘使用情况

### 2. **cleanup-system-disk.bat** - 紧急清理
紧急清理脚本，停止所有容器并清理所有缓存。

**使用方法：**
```bash
# 紧急清理系统盘
cleanup-system-disk.bat
```

**功能：**
- ✅ 停止所有容器
- ✅ 清理所有 Docker 构建缓存
- ✅ 清理未使用的镜像
- ✅ 清理未使用的卷
- ✅ 彻底清理（包含 filter "until=0s"）

### 3. **cleanup-docker.sh / cleanup-docker.ps1** - 常规清理
轻量级清理脚本，不停止容器，只清理未使用资源。

**使用方法：**
```bash
# PowerShell
.\cleanup-docker.ps1

# Bash (WSL/Linux)
./cleanup-docker.sh
```

## 📊 磁盘空间优化建议

### 1. **定期清理**
建议每次构建前或构建后都运行清理脚本，避免缓存累积。

### 2. **使用推荐的构建命令**
```bash
# 不推荐（会产生大量缓存）
docker-compose up --build

# 推荐（自动清理）
build-and-cleanup.bat
```

### 3. **监控磁盘使用**
定期检查 Docker 磁盘使用情况：
```bash
docker system df
```

输出示例：
```
TYPE            TOTAL     ACTIVE    SIZE      RECLAIMABLE
Images          3         3         898.5MB   0B (0%)
Containers      3         3         5.295MB   0B (0%)
Local Volumes   4         3         263.6MB   49.84MB (18%)
Build Cache     0         0         0B        0B
```

### 4. **构建前必清理**
养成好习惯：**构建前**和**构建后**都执行清理。

## 🎯 最佳实践

1. **日常开发**
   - 使用 `build-and-cleanup.bat` 进行构建
   - 避免使用 `docker-compose up --build`

2. **紧急清理**
   - 系统盘空间不足时使用 `cleanup-system-disk.bat`
   - 会停止所有运行中的容器

3. **定期维护**
   - 每周执行一次 `cleanup-system-disk.bat`
   - 保持 Docker 磁盘占用在合理范围

## 📝 脚本文件列表

| 文件名 | 用途 | 影响 |
|--------|------|------|
| `build-and-cleanup.bat` | 推荐构建脚本 | 自动清理 |
| `cleanup-system-disk.bat` | 紧急彻底清理 | 停止容器 |
| `cleanup-docker.ps1` | PowerShell清理 | 不停止容器 |
| `cleanup-docker.sh` | Bash清理 | 不停止容器 |

## ✅ 验证清理效果

清理完成后，运行以下命令验证：

```bash
docker system df
```

理想状态：
- **Build Cache**: 0B 或很小（< 1GB）
- **Images**: 只保留正在使用的镜像
- **Volumes**: 只保留正在使用的卷

## 🚨 注意事项

1. **不要频繁清理正在使用的缓存**
   - 会导致后续构建变慢
   - 只在需要时清理

2. **紧急清理前保存工作**
   - `cleanup-system-disk.bat` 会停止所有容器
   - 确保没有未保存的工作

3. **系统盘空间监控**
   - 定期检查 C:\ 盘空间
   - 及时清理避免系统崩溃
