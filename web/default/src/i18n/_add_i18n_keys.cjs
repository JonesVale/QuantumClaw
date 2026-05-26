/**
 * 脚本: 添加 Apps + Enterprise 页面缺少的 i18n 键
 * 用法: node _add_i18n_keys.cjs
 *
 * 需要添加的键:
 * - Apps: 9 descriptions + 4 categories
 * - Enterprise FEATURES: 6 titles + 6 descriptions (已用 t()，缺键值)
 * - Enterprise METRICS: 4 labels + 4 subs
 */
const fs = require('fs')
const path = require('path')

// ====== 新增的键（按语义化键名） ======
const APP_DESC_KEYS = {
  app_cursor_desc:       "AI-first code editor with multi-model support.",
  app_chatbox_desc:      "All-in-one AI desktop client supporting all major LLMs.",
  app_continue_desc:     "Open-source AI code assistant for VS Code & JetBrains.",
  app_lobechat_desc:     "Modern chat framework with plugin system and multi-model support.",
  app_openwebui_desc:    "Self-hosted WebUI for LLMs with QuantumClaw API integration.",
  app_dify_desc:         "Open-source LLM app development platform. Build and deploy AI applications.",
  app_fastgpt_desc:      "Knowledge-based Q&A system built on LLMs and vector databases.",
  app_cherry_desc:       "Desktop client for LLMs with support for multiple AI services.",
  app_aitoolkit_desc:    "Browser extension for ChatGPT, Gemini, Claude with QuantumClaw API support.",
}

const APP_CAT_KEYS = {
  app_category_Development: "Development",
  app_category_Chat:        "Chat",
  app_category_Platform:    "Platform",
  app_category_Tools:       "Tools",
}

// Enterprise 中已用 t() 的 FEATURES 键 - key 就是英文原文
const ENT_FEATURE_TITLES = {
  "SLA 99.99%": "SLA 99.99%",
  "Private Deployment": "Private Deployment",
  "Real-time Analytics": "Real-time Analytics",
  "Team & RBAC": "Team & RBAC",
  "Custom Integration": "Custom Integration",
  "Priority Support": "Priority Support",
}

const ENT_FEATURE_DESCS = {
  "Enterprise-grade reliability with multi-region failover and guaranteed uptime. Your AI infrastructure never goes dark.":
    "Enterprise-grade reliability with multi-region failover and guaranteed uptime. Your AI infrastructure never goes dark.",
  "Deploy on your infrastructure — VPC, on-premises, or air-gapped. Complete data isolation for regulated industries.":
    "Deploy on your infrastructure — VPC, on-premises, or air-gapped. Complete data isolation for regulated industries.",
  "Granular dashboards tracking every token, user, and cost center. Alerts, anomaly detection, and export to your BI tools.":
    "Granular dashboards tracking every token, user, and cost center. Alerts, anomaly detection, and export to your BI tools.",
  "Multi-team access control with API key rotation, usage quotas, per-model permissions, and full audit logging.":
    "Multi-team access control with API key rotation, usage quotas, per-model permissions, and full audit logging.",
  "SDKs for every language, webhook event streams, and dedicated middleware. Drop into your existing stack in hours.":
    "SDKs for every language, webhook event streams, and dedicated middleware. Drop into your existing stack in hours.",
  "Dedicated engineer with 15-minute response SLA. 24/7 coverage including on-call escalation and incident management.":
    "Dedicated engineer with 15-minute response SLA. 24/7 coverage including on-call escalation and incident management.",
}

// Enterprise METRICS - label & sub (需要加 t())
const ENT_METRIC_KEYS = {
  ent_metric_uptime_label:  "Uptime SLA",
  ent_metric_uptime_sub:    "Multi-region failover",
  ent_metric_models_label:  "AI Models",
  ent_metric_models_sub:    "Single unified API",
  ent_metric_response_label:"Response Time",
  ent_metric_response_sub:  "Priority support SLA",
  ent_metric_daily_label:   "Daily Tokens",
  ent_metric_daily_sub:     "Processed reliably",
}

// ====== 中文翻译 ======
const ZH_APP_DESC = {
  app_cursor_desc:       "支持多模型的 AI 优先代码编辑器。",
  app_chatbox_desc:      "一站式 AI 桌面客户端，支持所有主流大语言模型。",
  app_continue_desc:     "面向 VS Code 和 JetBrains 的开源 AI 代码助手。",
  app_lobechat_desc:     "支持插件系统和多模型的现代聊天框架。",
  app_openwebui_desc:    "集成 QuantumClaw API 的自托管 LLM Web 界面。",
  app_dify_desc:         "开源的 LLM 应用开发平台，轻松构建和部署 AI 应用。",
  app_fastgpt_desc:      "基于 LLM 和向量数据库的知识问答系统。",
  app_cherry_desc:       "支持多种 AI 服务的桌面客户端。",
  app_aitoolkit_desc:    "集成 QuantumClaw API 的 ChatGPT、Gemini、Claude 浏览器扩展。",
}

const ZH_APP_CAT = {
  app_category_Development: "开发",
  app_category_Chat:        "聊天",
  app_category_Platform:    "平台",
  app_category_Tools:       "工具",
}

const ZH_ENT_FEATURE_TITLES = {
  "SLA 99.99%": "SLA 99.99%",
  "Private Deployment": "私有部署",
  "Real-time Analytics": "实时分析",
  "Team & RBAC": "团队与权限",
  "Custom Integration": "自定义集成",
  "Priority Support": "优先支持",
}

const ZH_ENT_FEATURE_DESCS = {
  "Enterprise-grade reliability with multi-region failover and guaranteed uptime. Your AI infrastructure never goes dark.":
    "企业级可靠性，支持多区域故障转移和正常运行时间保证，您的 AI 基础架构永不掉线。",
  "Deploy on your infrastructure — VPC, on-premises, or air-gapped. Complete data isolation for regulated industries.":
    "在您的基础设施上部署——VPC、本地部署或物理隔离环境，为合规行业提供完全数据隔离。",
  "Granular dashboards tracking every token, user, and cost center. Alerts, anomaly detection, and export to your BI tools.":
    "精细仪表盘追踪每个 token、用户和成本中心，支持告警、异常检测和导出到 BI 工具。",
  "Multi-team access control with API key rotation, usage quotas, per-model permissions, and full audit logging.":
    "多团队访问控制，支持 API 密钥轮换、使用额度、按模型权限和完整审计日志。",
  "SDKs for every language, webhook event streams, and dedicated middleware. Drop into your existing stack in hours.":
    "支持所有语言 SDK、Webhook 事件流和专用中间件，数小时内即可集成到现有技术栈。",
  "Dedicated engineer with 15-minute response SLA. 24/7 coverage including on-call escalation and incident management.":
    "专属工程师，15 分钟响应 SLA，7×24 小时覆盖，含电话升级和事故管理。",
}

const ZH_ENT_METRIC = {
  ent_metric_uptime_label:   "正常运行时间",
  ent_metric_uptime_sub:     "多区域故障转移",
  ent_metric_models_label:   "AI 模型数",
  ent_metric_models_sub:     "统一 API 接入",
  ent_metric_response_label: "响应时间",
  ent_metric_response_sub:   "优先支持 SLA",
  ent_metric_daily_label:    "每日 Tokens",
  ent_metric_daily_sub:      "可靠处理",
}

// ====== 合并所有英文键 ======
const ALL_EN = {
  ...APP_DESC_KEYS,
  ...APP_CAT_KEYS,
  ...ENT_FEATURE_TITLES,
  ...ENT_FEATURE_DESCS,
  ...ENT_METRIC_KEYS,
}

const ALL_ZH = {
  ...ZH_APP_DESC,
  ...ZH_APP_CAT,
  ...ZH_ENT_FEATURE_TITLES,
  ...ZH_ENT_FEATURE_DESCS,
  ...ZH_ENT_METRIC,
}

// ====== 写入文件 ======
function addKeysToJson(filePath, newKeys) {
  const raw = fs.readFileSync(filePath, 'utf-8')
  const json = JSON.parse(raw)

  let added = 0, skipped = 0
  for (const [k, v] of Object.entries(newKeys)) {
    if (k in json) {
      skipped++
      continue
    }
    json[k] = v
    added++
  }

  const sorted = Object.keys(json).sort()
  const ordered = {}
  for (const k of sorted) {
    ordered[k] = json[k]
  }

  fs.writeFileSync(filePath, JSON.stringify(ordered, null, 2) + '\n', 'utf-8')
  console.log(`  ${filePath}: +${added}, skipped=${skipped}, total=${sorted.length}`)
}

addKeysToJson(path.join(__dirname, 'en.json'), ALL_EN)
addKeysToJson(path.join(__dirname, 'zh-CN.json'), ALL_ZH)

console.log('\nDone. Now update the frontend files and rebuild.')
