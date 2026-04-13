# 哪吒网络安全分析器

## 项目简介

**哪吒网络安全分析器**是一款面向授权渗透测试的红队辅助工具。它通过 **DeepSeek 大模型的函数调用能力**动态规划攻击路径，并自动执行 `nmap`、`sqlmap`、`nuclei`、`ffuf` 等常用安全工具。整个过程通过 **终端图形界面 (TUI)** 实时展示，用户可随时干预。

> 命名寓意：哪吒脚踏风火轮、手持火尖枪，象征**快速响应**与**精准打击**，正契合自动化渗透测试中"风驰电掣、直击弱点"的愿景。

## 作者信息

- **作者**: 钟智强
- **微信**: ctkqiang
- **电邮**: johnmelodymel@qq.com

## 核心功能

1. **AI 智能规划**

   - 输入目标 URL/IP 后，DeepSeek 分析并返回一组待执行的工具调用（`tool_calls`）。

2. **模块化工具执行器**

   - 支持 `nmap`、`sqlmap`、`nuclei`、`ffuf`、`subfinder`、`amass`、`httpx`、`commix`、`msfconsole`、`impacket`、`sliver-cli` 等。
   - 工具输出被解析为结构化数据，存入上下文。

3. **闭环反馈**

   - 每轮工具执行结果会**回传给 AI**，AI 据此决定是否追加扫描步骤。

4. **实时 TUI 面板**

   - 基于 Bubble Tea 构建，展示攻击计划、执行进度、实时日志。

5. **安全约束**
   - 自动拦截内网/政府域名等敏感目标。

## 系统架构

```
用户命令行 (-u target)
        │
        ▼
┌─────────────────┐     工具调用请求      ┌─────────────────┐
│   TUI 界面      │ ◄──────────────────► │   Orchestrator  │
│  (Bubble Tea)   │                      │   (状态机)       │
└─────────────────┘                      └────────┬────────┘
                                                  │
                                  ① 获取攻击计划   │  ② 返回工具列表
                                                  ▼
                                        ┌─────────────────┐
                                        │  DeepSeek 客户端 │
                                        │  (函数调用模式)  │
                                        └─────────────────┘
                                                  │
                                                  │ ③ 工具名称 + 参数
                                                  ▼
                                        ┌─────────────────┐
                                        │   工具注册表     │
                                        │ (nmap/sqlmap/...)│
                                        └────────┬────────┘
                                                 │
                                    ④ 执行       │
                                                 ▼
                                        ┌─────────────────┐
                                        │   执行器         │
                                        │ (调用外部命令)   │
                                        └────────┬────────┘
                                                 │
                                    ⑤ 结构化结果 │
                                                 ▼
                                        ┌─────────────────┐
                                        │   上下文存储     │
                                        │  (端口/服务/漏洞) │
                                        └─────────────────┘
                                                 │
                                    ⑥ 反馈结果   │
                                                 └──────► 回到 DeepSeek (下一轮)
```

## 技术栈

| 模块         | 技术方案                                                                                                         |
| ------------ | ---------------------------------------------------------------------------------------------------------------- |
| 编程语言     | Go 1.26+                                                                                                         |
| AI 接口      | DeepSeek API (函数调用)                                                                                          |
| TUI 框架     | [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss) |
| 端口扫描     | nmap                                                                                                             |
| 目录爆破     | ffuf                                                                                                             |
| SQL 注入检测 | sqlmap                                                                                                           |
| 漏洞扫描     | nuclei                                                                                                           |
| 子域名发现   | subfinder                                                                                                        |
| 资产发现     | amass                                                                                                            |
| HTTP 探测    | httpx                                                                                                            |
| 命令注入检测 | commix                                                                                                           |

## 项目结构

```
nezha_sec/
├── internal/
│   ├── api/              # DeepSeek API 客户端
│   │   └── deepseek.go
│   ├── executor/         # 工具执行器
│   │   ├── executor.go
│   │   ├── sqlmap_executor.go
│   │   ├── nmap_executor.go
│   │   ├── nuclei_executor.go
│   │   ├── ffuf_executor.go
│   │   ├── subfinder_executor.go
│   │   ├── amass_executor.go
│   │   ├── httpx_executor.go
│   │   ├── commix_executor.go
│   │   ├── msfconsole_executor.go
│   │   ├── impacket_executor.go
│   │   ├── sliver_cli_executor.go
│   │   └── ...
│   ├── model/            # 数据模型
│   │   ├── execution.go
│   │   └── tui.go
│   ├── orchestrator/     # 核心调度器
│   │   ├── orchestrator.go
│   │   └── workflow.go
│   ├── registry/         # 工具注册表
│   │   └── tool_registry.go
│   ├── security/         # 安全约束
│   │   └── constraints.go
│   ├── output/           # 输出美化
│   │   └── pretty_output.go
│   └── views/            # TUI 界面
│       ├── cli_views.go
│       ├── mcp_views.go
│       └── chat_views.go
├── .env                  # 环境配置文件
├── go.mod
├── go.sum
└── main.go
```

## 安装与配置

### 1. 克隆项目

```bash
git clone https://gitcode.com/ctkqiang_sr/NezhaSec.git
cd NezhaSec
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置 API 密钥

在 `.env` 文件中设置你的 DeepSeek API 密钥：

```env
APP_NAME=哪吒网络安全分析器
DEEPSEEK_API_KEY=your_api_key_here
```

或者通过环境变量设置：

```bash
export DEEPSEEK_API_KEY=your_api_key_here
```

### 4. 编译项目

```bash
go build -o nezha_sec .
```

## 使用方法

### 1. 基础扫描模式

```bash
./nezha_sec -u https://example.com
```

### 2. 红队工作流模式 (TUI 界面)

```bash
./nezha_sec -workflow
```

启动后输入目标 URL，程序会自动执行完整的渗透测试流程：

- 侦察阶段：子域名发现、资产发现、HTTP 探测、端口扫描
- 漏洞扫描阶段：漏洞扫描、目录爆破、SQL 注入检测
- 利用阶段：命令注入检测
- 权限提升阶段：权限提升模块搜索
- 横向移动阶段：横向移动探测
- 后利用阶段：后利用模块检查
- 报告阶段：生成完整报告

### 3. MCP 风格界面

```bash
./nezha_sec -mcp
```

### 4. 处理输出模式

```bash
./nezha_sec -process
```

## 支持的工具

### 侦察工具

- **subfinder**: 子域名发现
- **amass**: 资产发现
- **httpx**: HTTP 探测
- **nmap**: 端口扫描

### 漏洞扫描工具

- **nuclei**: 漏洞扫描
- **ffuf**: 目录爆破
- **sqlmap**: SQL 注入检测（支持 100+ 参数）

### 利用工具

- **commix**: 命令注入检测
- **msfconsole**: Metasploit 框架
- **impacket**: 横向移动
- **sliver-cli**: 后利用框架

## SQLMap 参数支持

SQLMap 执行器支持以下参数：

### 基础参数

- `url`: 目标 URL
- `risk`: 风险等级 (1-3)
- `level`: 检测等级 (1-5)
- `threads`: 线程数

### HTTP 相关

- `cookie`: Cookie 数据
- `data`: POST 数据
- `method`: HTTP 方法
- `headers`: 自定义请求头
- `user-agent`: User-Agent
- `referer`: Referer
- `proxy`: 代理设置
- `tor`: 使用 Tor 网络

### 注入优化

- `dbms`: 指定数据库类型
- `tamper`: 使用 tamper 脚本
- `technique`: 指定注入技术
- `union-cols`: UNION 查询列数
- `prefix`/`suffix`: 前缀/后缀

### 数据提取

- `dump`: 转储数据
- `dump-all`: 转储所有数据
- `dbs`: 枚举数据库
- `tables`: 枚举表
- `columns`: 枚举列
- `db`/`tbl`/`col`: 指定数据库/表/列

### 高级功能

- `sql-query`: 执行 SQL 查询
- `os-cmd`: 执行操作系统命令
- `os-shell`: 获取操作系统 shell
- `file-read`: 读取文件
- `file-write`: 写入文件

## TUI 快捷键

| 按键             | 功能               |
| ---------------- | ------------------ |
| `Enter`          | 确认输入/开始执行  |
| `P`              | 暂停/恢复工作流    |
| `C`              | 重新开始（完成后） |
| `Ctrl+C` / `Esc` | 退出程序           |

## 安全声明

- 本工具**仅限授权测试**使用（例如自有系统、CTF 平台、客户授权的渗透项目）。
- 内置**目标白名单机制**，默认阻止扫描 `.gov`、`.mil` 及内网地址。
- 所有操作日志均带有时间戳，便于审计。
- 使用者须遵守当地法律法规，**未经授权扫描他人系统属违法行为**。

## 许可证

本项目仅供学习和授权测试使用，请遵守相关法律法规。

## 联系方式

- **作者**: 钟智强
- **微信**: ctkqiang
- **电邮**: johnmelodymel@qq.com

---

**免责声明**: 本工具仅用于授权的安全测试和研究目的。使用本工具进行未经授权的访问或攻击是违法的。作者不对任何非法使用承担责任。
