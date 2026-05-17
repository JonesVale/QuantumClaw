export interface NewsSource {
  name: string
  url: string
  lang: string
  description: string
  category: string
}

export const newsSources: NewsSource[] = [
  // ── 中文 ──────────────────────────────────────────────
  {
    name: '机器之心',
    url: 'https://www.jiqizhixin.com',
    lang: 'zh',
    description: '专注于人工智能领域的专业媒体，提供深度技术报道、行业分析与前沿动态追踪',
    category: 'tech',
  },
  {
    name: '量子位',
    url: 'https://www.qbitai.com',
    lang: 'zh',
    description: '关注 AI 技术落地与产业应用，覆盖大模型、自动驾驶、机器人等热门方向',
    category: 'tech',
  },
  {
    name: '36氪 AI',
    url: 'https://36kr.com/information/AI/',
    lang: 'zh',
    description: '36氪旗下 AI 频道，聚合人工智能领域的创业、投资与前沿技术资讯',
    category: 'tech',
  },
  {
    name: '雷锋网',
    url: 'https://www.leiphone.com/category/ai',
    lang: 'zh',
    description: '聚焦 AI 科技与产业，涵盖学术突破、商业动态及智能硬件评测',
    category: 'tech',
  },
  // ── English ────────────────────────────────────────────
  {
    name: 'OpenAI Blog',
    url: 'https://openai.com/blog',
    lang: 'en',
    description: 'Official blog from OpenAI — product launches, research breakthroughs, and safety updates',
    category: 'official',
  },
  {
    name: 'Anthropic',
    url: 'https://www.anthropic.com/blog',
    lang: 'en',
    description: 'Official Anthropic blog covering Claude model updates, alignment research, and company news',
    category: 'official',
  },
  {
    name: 'Google AI',
    url: 'https://ai.googleblog.com',
    lang: 'en',
    description: 'Google\'s official AI blog — Gemini, PaLM, DeepMind research, and product announcements',
    category: 'official',
  },
  {
    name: 'MIT Tech Review',
    url: 'https://www.technologyreview.com/topic/artificial-intelligence/',
    lang: 'en',
    description: 'In-depth reporting on AI\'s impact on society, business, and science from MIT',
    category: 'tech',
  },
  {
    name: 'ArXiv',
    url: 'https://arxiv.org',
    lang: 'en',
    description: 'Open-access preprint repository for cutting-edge AI/ML research papers',
    category: 'research',
  },
  {
    name: 'Hugging Face',
    url: 'https://huggingface.co/blog',
    lang: 'en',
    description: 'Community blog from Hugging Face — model releases, datasets, and open-source AI tools',
    category: 'community',
  },
  {
    name: 'Reddit AI',
    url: 'https://www.reddit.com/r/artificial/',
    lang: 'en',
    description: 'Reddit community discussing AI news, discussions, and emerging trends',
    category: 'community',
  },
]
