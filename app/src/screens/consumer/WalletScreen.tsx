/**
 * Consumer WalletScreen - Balance & Top-up
 *
 * Data sources:
 *  - getSelf() → balance
 *  - getTransactions() → history
 *  - createTopUpOrder() → payment
 */

import React, { useEffect, useState } from 'react'
import {
  View, Text, StyleSheet, ScrollView, TouchableOpacity, RefreshControl,
  Alert, ActivityIndicator,
} from 'react-native'
import { getSelf, getTransactions, createTopUpOrder, Transaction } from '../../lib/api-extended'
import { useAuthStore } from '../../stores/auth-store'

const TOPUP_AMOUNTS = [1000, 5000, 10000, 50000, 100000, 500000]

export default function WalletScreen() {
  const user = useAuthStore((s) => s.user)
  const setUser = useAuthStore((s) => s.setUser)
  const [txns, setTxns] = useState<Transaction[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [topupLoading, setTopupLoading] = useState<number | null>(null)

  const fetchData = async () => {
    try {
      const [selfRes, txnRes] = await Promise.all([
        getSelf(),
        getTransactions(1, 50),
      ])
      if (selfRes.success && selfRes.data) {
        setUser({
          id: selfRes.data.id, username: selfRes.data.username,
          display_name: selfRes.data.display_name, email: selfRes.data.email,
          role: selfRes.data.role, status: selfRes.data.status,
          quota: selfRes.data.quota, used_quota: selfRes.data.used_quota,
          request_count: selfRes.data.request_count,
        })
      }
      if (txnRes.success && txnRes.data) setTxns(txnRes.data)
    } catch { /* ignore */ }
    finally { setLoading(false); setRefreshing(false) }
  }

  useEffect(() => { fetchData() }, [])

  const handleTopup = async (amount: number) => {
    setTopupLoading(amount)
    try {
      const res = await createTopUpOrder(amount, 'alipay')
      if (res.success && res.data) {
        Alert.alert(
          '充值订单已创建',
          `金额: ¥${(amount / 100).toFixed(2)}\n订单号: ${res.data.order_no}`,
        )
        fetchData()
      } else {
        Alert.alert('充值失败', res.message || '请重试')
      }
    } catch (e: any) {
      Alert.alert('充值失败', e.message || '网络错误')
    }
    finally { setTopupLoading(null) }
  }

  const balance = user?.balance != null ? `¥${(user.balance / 100).toFixed(2)}` : '—'

  if (loading) {
    return <View style={s.center}><ActivityIndicator size="large" color="#d97706" /></View>
  }

  return (
    <ScrollView
      style={s.container}
      contentContainerStyle={s.content}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={() => { setRefreshing(true); fetchData() }} tintColor="#d97706" />}
    >
      {/* Balance Card */}
      <View style={s.balanceCard}>
        <Text style={s.balanceLabel}>💰 账户余额</Text>
        <Text style={s.balanceAmount}>{balance}</Text>
        <Text style={s.balanceHint}>
          已用配额: {user?.used_quota != null ? `${(user.used_quota / 1000000).toFixed(1)}M` : '—'}
        </Text>
      </View>

      {/* Quick Top-up */}
      <View style={s.topupSection}>
        <Text style={s.sectionTitle}>⚡ 快速充值</Text>
        <View style={s.amountGrid}>
          {TOPUP_AMOUNTS.map((amt) => (
            <TouchableOpacity
              key={amt}
              style={s.amountBtn}
              onPress={() => handleTopup(amt)}
              disabled={topupLoading !== null}
            >
              {topupLoading === amt ? (
                <ActivityIndicator color="#d97706" />
              ) : (
                <Text style={s.amountText}>¥{(amt / 100).toFixed(0)}</Text>
              )}
            </TouchableOpacity>
          ))}
        </View>
      </View>

      {/* Transactions */}
      <View style={s.txnSection}>
        <Text style={s.sectionTitle}>🔄 交易记录</Text>
        {txns.length === 0 && (
          <Text style={s.empty}>暂无交易记录</Text>
        )}
        {txns.map((t) => (
          <View key={t.id} style={s.txnItem}>
            <View style={s.txnLeft}>
              <Text style={s.txnDesc}>{t.description || '交易'}</Text>
              <Text style={s.txnTime}>
                {t.created_time ? new Date(t.created_time * 1000).toLocaleString() : ''}
              </Text>
            </View>
            <Text style={[
              s.txnAmount,
              { color: t.amount > 0 ? '#16a34a' : '#dc2626' }
            ]}>
              {t.amount > 0 ? '+' : ''}{t.amount}
            </Text>
          </View>
        ))}
      </View>
    </ScrollView>
  )
}

const s = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f9fafb' },
  content: { padding: 16, paddingBottom: 32 },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: '#f9fafb' },

  balanceCard: { backgroundColor: '#fff', borderRadius: 20, padding: 24, alignItems: 'center', marginBottom: 16, borderWidth: 1, borderColor: '#f3f0ea' },
  balanceLabel: { fontSize: 14, color: '#6b7280', marginBottom: 8 },
  balanceAmount: { fontSize: 36, fontWeight: '700', color: '#d97706', marginBottom: 4 },
  balanceHint: { fontSize: 12, color: '#9ca3af' },

  topupSection: { marginBottom: 16 },
  sectionTitle: { fontSize: 16, fontWeight: '600', color: '#111827', marginBottom: 12 },
  amountGrid: { flexDirection: 'row', flexWrap: 'wrap', gap: 8 },
  amountBtn: {
    width: '31%', backgroundColor: '#fffbeb', borderRadius: 14, padding: 16,
    alignItems: 'center', borderWidth: 1, borderColor: '#fde68a',
  },
  amountText: { fontSize: 16, fontWeight: '600', color: '#d97706' },

  txnSection: { marginBottom: 16 },
  txnItem: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', backgroundColor: '#fff', borderRadius: 12, padding: 14, marginBottom: 8 },
  txnLeft: { flex: 1 },
  txnDesc: { fontSize: 14, color: '#374151', fontWeight: '500' },
  txnTime: { fontSize: 11, color: '#9ca3af', marginTop: 2 },
  txnAmount: { fontSize: 16, fontWeight: '700' },
  empty: { color: '#9ca3af', fontSize: 13, textAlign: 'center', paddingVertical: 20 },
})
