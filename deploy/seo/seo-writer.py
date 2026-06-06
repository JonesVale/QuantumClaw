#!/usr/bin/env python3
"""
QuantumClaw SEO 文章生成器 - 每周执行
每周一自动生成 3-5 篇原创 SEO 文章，围绕 AI API 长尾关键词

调用方式: python3 /path/to/seo-writer.py
输出: 通过 API 写入 rss_articles 表
"""

import json
import random
import datetime
import sys
import os

# ── 长尾关键词库（按主题分类） ──
TOPICS = {
    "api-guide": [
        "API Token 选购避坑指南",
        "AI API 接口限流解决方案",
        "Codex 怎么注册|官方注册教程",
        "DeepSeek V4 接口对接最佳实践",
        "Claude API Key 申请与配置教程",
        "OpenAI API 代理设置与网络优化",
        "多家 AI API 聚合平台选型对比",
        "AI API 调用费用控制与优化策略",
    ],
    "model-tutorial": [
        "GPT-4o 接口调用参数详解",
        "Claude Sonnet 4 API 接入教程",
        "DeepSeek V4 模型能力全解析",
        "Gemini 3.0 Pro API 使用指南",
        "O1 Mini 推理模型适用场景分析",
        "Qwen 2.5 VL 视觉模型接口教程",
        "Llama 3 开源模型 API 部署方案",
    ],
    "dev-tips": [
        "Python 调用 AI API 完整代码示例",
        "TypeScript/Node.js AI API 集成指南",
        "WebSocket 流式响应接入最佳实践",
        "API Key 安全管理与轮换策略",
        "AI 模型调用日志分析与监控方案",
        "多模型负载均衡与故障转移配置",
    ],
    "platform": [
        "AI API 网关搭建完整教程",
        "渠道商 API 分销平台运营指南",
        "企业 AI API 统一管理解决方案",
        "API 调用量预测与成本预算管理",
        "AI 模型价格对比与选型建议",
        "2026年 AI API 发展趋势分析",
    ],
}

# ── 文章模版 ──
TEMPLATES = {
    "guide": """# {title}

## 为什么这很重要？

在当今 AI 快速发展的时代，{topic_desc} 成为开发者和企业关注的焦点。本文将深入分析 {focus}，帮助你做出最佳决策。

## {section1}

{content1}

## {section2}

{content2}

## 最佳实践

{best_practice}

## 常见问题

**Q: {q1}**
A: {a1}

**Q: {q2}**
A: {a2}

## 总结

{summary}

---

*本文由 QuantumClaw AI API 聚合平台原创发布。访问 qscl.link 获取更多 AI API 资讯。*""",
}

# ── 内容填充器（每次生成不同内容） ──
def fill_template(title: str, topic_key: str) -> dict:
    today = datetime.date.today()
    week_num = today.isocalendar()[1]
    
    articles = []
    keywords = TOPICS.get(topic_key, TOPICS["api-guide"])
    
    # 选 3-5 篇
    count = min(random.randint(3, 5), len(keywords))
    selected = random.sample(keywords, count)
    
    for kw in selected:
        parts = kw.split("|")
        main_title = parts[0]
        subtitle = parts[1] if len(parts) > 1 else ""
        
        content = f"""
<div class="seo-article">
<p>随着 AI 技术的快速发展，<strong>{main_title}</strong> 成为越来越多开发者关注的话题。QuantumClaw 作为领先的 AI API 聚合平台，接入全球 400+ 大模型，为用户提供统一的 API 接入体验。</p>

<p>本教程将详细介绍如何快速上手 {main_title}，包括环境配置、API 调用示例、常见问题排查等内容。无论你是初学者还是有经验的开发者，都能从中获得实用价值。</p>

<h3>快速开始</h3>
<ol>
<li>注册 QuantumClaw 账号（qscl.link）</li>
<li>创建 API Key，获取专属访问凭证</li>
<li>选择目标模型，复制对应的 API 端点</li>
<li>集成到你的代码中，开始调用 AI 能力</li>
</ol>

<h3>代码示例</h3>
<pre><code>
# Python 示例 - 使用 OpenAI 兼容接口
import openai

client = openai.OpenAI(
    api_key="你的_API_KEY",
    base_url="https://api.qscl.link/v1"
)

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{{"role": "user", "content": "Hello!"}}]
)
print(response.choices[0].message.content)
</code></pre>

<h3>注意事项</h3>
<ul>
<li>确保 API Key 有足够额度</li>
<li>根据并发需求选择合适的套餐</li>
<li>开启监控报警，及时发现问题</li>
</ul>
</div>
"""
        
        articles.append({
            "title": f"{main_title} - {today.year}年{week_num}周更新",
            "content": content,
            "author": "QuantumClaw Team",
            "language": "zh",
            "source": "seo",
        })
    
    return articles


# ── 主函数 ──
def main():
    print(f"[SEO Writer] 开始生成第 {datetime.date.today().isocalendar()[1]} 周内容")
    
    all_articles = []
    
    # 每周轮换不同主题
    topics = list(TOPICS.keys())
    random.shuffle(topics)
    
    # 选 2 个主题，每个生成 2-3 篇
    selected_topics = topics[:2]
    for topic in selected_topics:
        articles = fill_template(topic, topic)
        all_articles.extend(articles)
    
    print(f"  生成了 {len(all_articles)} 篇文章")
    for art in all_articles:
        print(f"    - {art['title']}")
    
    # 直接输出 JSON（cron 任务捕获并调用 API 写入）
    output = json.dumps(all_articles, ensure_ascii=False, indent=2)
    
    # 写入文件供调试
    output_dir = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "seo")
    os.makedirs(output_dir, exist_ok=True)
    output_path = os.path.join(output_dir, f"articles-week-{datetime.date.today().isocalendar()[1]}.json")
    with open(output_path, "w", encoding="utf-8") as f:
        f.write(output)
    
    print(f"\n  已保存到: {output_path}")
    print(output)  # stdout 输出给 cron 捕获

if __name__ == "__main__":
    main()
