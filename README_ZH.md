# mop

<p align="center">
  <img src="assets/Headpic.png" alt="mop" width="600">
</p>

[English](README.md) | 中文文档

AI 编程工具会在你电脑各处留下几百兆的缓存、会话和凭证——卸载后也不清理。mop 一个界面清点并安全清理所有这些垃圾。

## 特性

- 🚀 **极速启动**：单二进制 <10MB，启动 <50ms
- 🎨 **TUI 交互**：基于 Bubble Tea 的终端界面
- 🗑️ **安全删除**：默认移到废纸篓，可随时恢复
- ⏱️ **时间过滤**：按时间范围筛选（全部 / 3天内 / 7天内 / 30天内）
- 🔥 **大文件标记**：超过 100MB 的项自动高亮
- 📝 **白名单**：按 W 键快速加入/移出白名单
- ⚙️ **可配置**：Manage Tools 按需启用/禁用扫描器
- 🤖 **脚本友好**：支持非交互模式（`mop clean`）

## 安装

### 快速安装（推荐）

**Apple Silicon (M1/M2/M3/M4)**

```bash
curl -L https://github.com/Do-ooo/mop/releases/latest/download/mop-darwin-arm64 -o mop && chmod +x mop && sudo mv mop /usr/local/bin/
```

**Intel Mac**

```bash
curl -L https://github.com/Do-ooo/mop/releases/latest/download/mop-darwin-amd64 -o mop && chmod +x mop && sudo mv mop /usr/local/bin/
```

安装后直接运行：
```bash
mop
```

### 源码编译

```bash
git clone https://github.com/Do-ooo/mop.git
cd mop
go build -o mop .
```

## 使用

### 交互模式（默认）

```bash
mop
```

进入主菜单，选择：
- **Analyze** — Regular 模式，扫描安全项（缓存/日志/临时文件）
- **Deep Analyze** — Deep 模式，扫描全部项（含会话历史，删了不可恢复）
- **Manage Tools** — 启用/禁用各工具的扫描器
- **About** — 关于信息

### 快捷键（选择界面）

| 键 | 功能 |
|----|------|
| `↑/↓` 或 `j/k` | 上下导航 |
| `Space` | 选中/取消选中 |
| `a` | 全选 |
| `i` | 反选 |
| `w` | 加入/移出白名单 |
| `r` | 重新扫描 |
| `t` | 切换时间过滤（全部→3天内→7天内→30天内） |
| `d` | 切换删除模式（Trash/Delete） |
| `Enter` | 开始清理（Deep 模式需二次确认） |
| `q` | 返回菜单 / 退出 |

### 命令行模式

#### 扫描（仅查看）

```bash
mop scan
```

#### 一键清理

```bash
# 默认移到废纸篓
mop clean

# 预览（不删除）
mop --dry-run

# 永久删除（跳过废纸篓）
mop clean --delete
```

#### 自更新

```bash
mop update
```

检查并下载最新版本，自动替换当前二进制。

### 支持的工具

| 工具 | 类型 |
|------|------|
| Trae | CLI + Desktop |
| WorkBuddy | CLI + Desktop |
| Cursor | Desktop |
| Windsurf | Desktop |
| VS Code | Desktop |
| Codex | CLI + Desktop |
| Claude Code | CLI |
| CodeBuddy | CLI + Desktop |
| Qoder | CLI + Desktop |
| OpenCode | CLI |
| Continue | CLI |
| Gemini | CLI + Desktop |
| Aider | CLI |
| Copilot | CLI |
| Codeium | CLI |
| JetBrains | Desktop |
| Augment | CLI |
| Supermaven | CLI |
| GitHub CLI | CLI |
| MiMo Code | CLI |

## 配置文件

配置存储在 `~/.config/mop/` 目录下：

- `config.json` — 全局配置（删除模式等）
- `whitelist.json` — 白名单路径
- `enabled_scanners.json` — 启用的扫描器

## 架构

```
main.go          # 入口，子命令分发
scanner/         # 扫描器，每个工具一个文件
cleaner/         # 清理器（文件删除 / 废纸篓）
tui/             # TUI 界面
config/          # 配置管理
whitelist/       # 白名单管理
update/          # 自更新逻辑
```
