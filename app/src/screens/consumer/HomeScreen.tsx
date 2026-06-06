/**
 * Consumer HomeScreen - Dashboard
 *
 * Data sources:
 *  - getSelf()         → user profile (balance, role, quota)
 *  - getDashboardData()→ usage stats, daily trends
 *  - getTokens()       → API key count for quick overview
 */

import React, { useEffect, useState, useCallback } from 'react'
import {
  View, Text, StyleSheet, ScrollView, TouchableOpacity, RefreshControl, ActivityIndicator,
} from 'react-native'
import { useNavigation } from '@react-navigation/native'
import { getSelf, getDashboardData, getTokens, DashboardData, Token } from '../../lib/api-extended'
import { useAuthStore } from '../../stores/auth-store'

type StatBoxProps = {
  icon: string; value: string; label: string; color?: string
}

function StatBox({ icon, value, label, color }: StatBoxProps) {
  return (
    <View style={s.statBox}>
      <Text style={s.statIcon}>{icon}</Text>
      <Text style={[s.statValue, color ? { color } : {}]}>{value}</Text>
      <Text style={s.statLabel}>{label}</Text>
    </View>
  )
}

export default function HomeScreen() {
  const nav = useNavigation()
  const user = useAuthStore((s) => s.user)
  const setUser = useAuthStore((s) => s.setUser)

  const [dashboard, setDashboard] = useState<DashboardData | null>(null)
  const [keys, setKeys] = useState<Token[]>([])
  const [refreshing, setRefreshing] = useState(false)
  const [loading, setLoading] = useState(true)

  const fetchData = useCallback(async () => {
    try {
      const [selfRes, dashRes, tokenRes] = await Promise.all([
        getSelf(),
        getDashboardData(),
        getTokens(),
      ])
      if (selfRes.success && selfRes.data) {
        const d = selfRes.data
        setUser({
          id: d.id, username: d.username, display_name: d.display_name,
          email: d.email, role: d.role, status: d.status,
          quota: d.quota, used_quota: d.used_quota, request_count: d.request_count,
        })
      }
      if (dashRes.success && dashRes.data) setDashboard(dashRes.data)
      if (tokenRes.success && tokenRes.data) setKeys(tokenRes.data)
    } catch { /* errors handled by api interceptor */ }
    finally { setLoading(false); setRefreshing(false) }
  }, [])

  useEffect(() => { fetchData() }, [])

  const onRefresh = () => { setRefreshing(true); fetchData() }

  const nickname = user?.display_name || user?.username || '用户'
  const balance = dashboard ? `¥${(dashboard.balance / 100).toFixed(2)}` : '—'
  const usedQuota = dashboard ? (dashboard.used_quota / 1000000).toFixed(1) + 'M' : '—'
  const totalQuota = dashboard ? (dashboard.quota / 1000000).toFixed(1) + 'M' : '—'
  const quotaPct = dashboard && dashboard.quota > 0
    ? Math.round((dashboard.used_quota / dashboard.quota) * 100) : 0

  if (loading) {
    return (
      <View style={s.center}>
        <ActivityIndicator size="large" color="#d97706" />
      </View>
    )
  }

  return (
    <ScrollView
      style={s.container}
      contentContainerStyle={s.content}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} tintColor="#d97706" />}
    >
      {/* Header */}
      <View style={s.header}>
        <View>
          <Text style={s.greeting}>👋 你好, {nickname}</Text>
          <Text style={s.role}>
            {user?.role === 100 ? '超级管理员' : user?.role === 10 ? '管理员' : '个人用户'}
          </Text>
        </View>
        <View style={s.avatarBox}>
          <Text style={s.avatarText}>{nickname[0]}</Text>
        </View>
      </View>

      {/* Stats Cards */}
      <View style={s.statsRow}>
        <StatBox icon="🔑" value={`${keys.length}`} label="API Key" />
        <StatBox icon="💰" value={balance} label="余额" color="#16a34a" />
        <StatBox icon="⚡" value={`${usedQuota}/${totalQuota}`} label="本月用量" />
      </View>

      {/* Quota Progress */}
      <View style={s.quotaCard}>
        <Text style={s.quotaTitle}>📊 配额使用率</Text>
        <View style={s.quotaBar}>
          <View style={[s.quotaFill, { width: `${Math.min(quotaPct, 100)}%` }]} />
        </View>
        <Text style={s.quotaText}>{quotaPct}% 已使用</Text>
      </View>

      {/* Quick Actions */}
      <View style={s.quickActions}>
        <TouchableOpacity style={s.actionBtn} onPress={() => nav.navigate('Keys' as never)}>
          <Text style={s.actionIcon}>🔑</Text>
          <Text style={s.actionLabel}>创建Key</Text>
        </TouchableOpacity>
        <TouchableOpacity style={s.actionBtn} onPress={() => nav.navigate('Wallet' as never)}>
          <Text style={s.actionIcon}>💰</Text>
          <Text style={s.actionLabel}>充值</Text>
        </TouchableOpacity>
        <TouchableOpacity style={s.actionBtn} onPress={() => {
          // Quick test: try calling /api/status
          fetch('/api/status').then(r => r.json()).then(d => {
            alert(`API Status: ${JSON.stringify(d)}`)
          }).catch(e => alert(`Error: ${e.message}`))
        }}>
          <Text style={s.actionIcon}>🧪</Text>
          <Text style={s.actionLabel}>调试</Text>
        </TouchableOpacity>
        <TouchableOpacity style={s.actionBtn} onPress={() => nav.navigate('Invite' as never)}>
          <Text style={s.actionIcon}>📨</Text>
          <Text style={s.actionLabel}>邀请</Text>
        </TouchableOpacity>
      </View>

      {/* Recent API Keys */}
      <View style={s.card}>
        <View style={s.cardHeader}>
          <Text style={s.cardTitle}>🔑 最近Key</Text>
          <TouchableOpacity onPress={() => nav.navigate('Keys' as never)}>
            <Text style={s.cardAction}>查看全部 →</Text>
          </TouchableOpacity>
        </View>
        {keys.slice(0, 3).map((k) => (
          <View key={k.id} style={s.keyItem}>
            <Text style={s.keyName}>{k.name || '未命名'}</Text>
            <Text style={s.keyVal}>
              {k.key ? k.key.substring(0, 8) + '****' : '***'}
            </Text>
            <Text style={s.keyQuota}>
              {k.remain_quota != null
                ? `${(k.remain_quota / 1000000).toFixed(1)}M`
                : (k.unlimited_quota ? '∞' : '—')}
            </Text>
          </View>
        ))}
        {keys.length === 0 && (
          <Text style={s.empty}>暂无 API Key，点击上方创建</Text>
        )}
      </View>

      {/* Daily Usage - placeholder for chart */}
      {dashboard?.daily_usage && dashboard.daily_usage.length > 0 && (
        <View style={s.card}>
          <Text style={s.cardTitle}>📈 最近调用</Text>
          {dashboard.daily_usage.slice(-5).map((d, i) => (
            <View key={i} style={s.usageRow}>
              <Text style={s.usageDate}>{d.date}</Text>
              <Text style={s.usageVal}>{(d.amount / 1000).toFixed(0)}K tokens</Text>
            </View>
          ))}
        </View>
      )}

      {/* Recent Transactions - latest 3 */}
      <View style={s.card}>
        <Text style={s.cardTitle}>🔄 最近流水</Text>
        {dashboard?.model_usage?.slice(0, 3).map((m, i) => (
          <View key={i} style={s.usageRow}>
            <Text style={s.usageDate}>🧠 {m.model}</Text>
            <Text style={s.usageVal}>{(m.quota / 1000).toFixed(0)}K</Text>
          </View>
        ))}
        {(!dashboard?.model_usage || dashboard.model_usage.length === 0) && (
          <Text style={s.empty}>暂无调用记录</Text>
        )}
      </View>
    </ScrollView>
  )
}

const s = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f9fafb' },
  content: { padding: 16, paddingBottom: 32 },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: '#f9fafb' },

  header: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 },
  greeting: { fontSize: 22, fontWeight: '700', color: '#111827' },
  role: { fontSize: 13, color: '#9ca3af', marginTop: 2 },
  avatarBox: { width: 44, height: 44, borderRadius: 22, backgroundColor: '#d97706', justifyContent: 'center', alignItems: 'center' },
  avatarText: { fontSize: 18, fontWeight: '700', color: '#fff' },

  statsRow: { flexDirection: 'row', gap: 8, marginBottom: 12 },
  statBox: { flex: 1, backgroundColor: '#fff', borderRadius: 14, padding: 14, alignItems: 'center', borderWidth: 1, borderColor: '#f3f0ea' },
  statIcon: { fontSize: 20, marginBottom: 4 },
  statValue: { fontSize: 18, fontWeight: '700', color: '#d97706', marginBottom: 2 },
  statLabel: { fontSize: 11, color: '#6b7280' },

  quotaCard: { backgroundColor: '#fffbeb', borderRadius: 16, padding: 16, marginBottom: 16, borderWidth: 1, borderColor: '#fde68a' },
  quotaTitle: { fontSize: 14, fontWeight: '600', color: '#92400e', marginBottom: 8 },
  quotaBar: { height: 8, backgroundColor: '#fef3c7', borderRadius: 4, overflow: 'hidden' },
  quotaFill: { height: 8, backgroundColor: '#d97706', borderRadius: 4 },
  quotaText: { fontSize: 11, color: '#92400e', marginTop: 6, textAlign: 'right' },

  quickActions: { flexDirection: 'row', gap: 8, marginBottom: 16 },
  actionBtn: { flex: 1, backgroundColor: '#fff', borderRadius: 14, padding: 14, alignItems: 'center', borderWidth: 1, borderColor: '#f3f0ea' },
  actionIcon: { fontSize: 24, marginBottom: 4 },
  actionLabel: { fontSize: 12, fontWeight: '500', color: '#374151' },

  card: { backgroundColor: '#fff', borderRadius: 16, padding: 16, marginBottom: 16 },
  cardHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 },
  cardTitle: { fontSize: 15, fontWeight: '600', color: '#111827', marginBottom: 12 },
  cardAction: { fontSize: 13, color: '#d97706', fontWeight: '500' },

  keyItem: { flexDirection: 'row', alignItems: 'center', paddingVertical: 10, borderBottomWidth: 1, borderBottomColor: '#f3f0ea' },
  keyName: { flex: 1, fontSize: 14, fontWeight: '500', color: '#111827' },
  keyVal: { fontSize: 12, color: '#9ca3af', marginRight: 8 },
  keyQuota: { fontSize: 12, color: '#6b7280', fontWeight: '500' },
  empty: { fontSize: 13, color: '#9ca3af', textAlign: 'center', paddingVertical: 20 },

  usageRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', paddingVertical: 8, borderBottomWidth: 1, borderBottomColor: '#f3f0ea' },
  usageDate: { fontSize: 13, color: '#374151' },
  usageVal: { fontSize: 13, color: '#6b7280', fontWeight: '500' },
})
