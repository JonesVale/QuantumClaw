/**
 * Consumer InviteScreen - 完整邀请页
 *
 * 全链路流程：
 *   1. 读取协议 → 用户同意
 *   2. 读取通讯录 (expo-contacts)
 *   3. 创建 MessageJob (后端)
 *   4. 启动发送引擎 (send-engine.ts)
 *   5. 实时进度显示
 *   6. 支持暂停/恢复/取消
 *   7. 断点续传
 */

import React, { useState, useEffect, useRef } from 'react'
import {
  View, Text, StyleSheet, ScrollView, TouchableOpacity, Alert,
  ActivityIndicator,
} from 'react-native'
import { useAuthStore } from '../../stores/auth-store'
import { getActiveAgreement, createMessageJob, getMessageJobs, MessageJob } from '../../lib/api-extended'
import {
  startSendJob, readContacts, setProgressCallback,
  pauseSend, resumeSend, cancelSend,
  SendProgress, Contact,
} from '../../lib/send-engine'

export default function InviteScreen() {
  const user = useAuthStore((s) => s.user)
  const [showAgreement, setShowAgreement] = useState(false)
  const [agreement, setAgreement] = useState<{ version: string; content: string } | null>(null)
  const [loading, setLoading] = useState(false)
  const [progress, setProgress] = useState<SendProgress | null>(null)
  const [contacts, setContacts] = useState<Contact[]>([])
  const [lastJob, setLastJob] = useState<MessageJob | null>(null)

  // Register progress callback
  useEffect(() => {
    setProgressCallback((p) => setProgress(p))
    return () => setProgressCallback(null)
  }, [])

  // ── 读取通讯录 + 创建 Job + 开始发送 ──
  const handleStart = async () => {
    try {
      // 1. Get agreement
      const agreeRes = await getActiveAgreement()
      if (!agreeRes.success || !agreeRes.data) {
        Alert.alert('提示', '暂无可用协议')
        return
      }
      setAgreement(agreeRes.data)
      setShowAgreement(true)
    } catch {
      Alert.alert('错误', '获取协议失败')
    }
  }

  // ── 用户同意协议 ──
  const handleAgree = async () => {
    setShowAgreement(false)
    setLoading(true)

    try {
      // 2. Read contacts
      const allContacts = await readContacts()
      if (allContacts.length === 0) {
        Alert.alert('提示', '通讯录中未找到有效联系人')
        return
      }

      setContacts(allContacts)

      // 3. Create message job
      const jobRes = await createMessageJob({
        channel: 'sms',
        total_targets: allContacts.length,
        agreement_version: agreement?.version || 'v1.0',
      })

      if (!jobRes.success || !jobRes.data) {
        Alert.alert('创建任务失败', jobRes.message)
        return
      }

      const job = jobRes.data
      setLastJob(job)

      // 4. Start send engine (non-blocking)
      startSendJob(job.id, allContacts).catch((err) => {
        Alert.alert('发送错误', err.message)
      })
    } catch (e: any) {
      if (e.message?.includes('permission')) {
        Alert.alert('权限被拒绝', '请在系统设置中允许访问通讯录')
      } else {
        Alert.alert('错误', e.message || '操作失败')
      }
    } finally {
      setLoading(false)
    }
  }

  const nickname = user?.display_name || user?.username || '好友'
  const affCode = 'ABC123' // Will be dynamically fetched

  return (
    <ScrollView style={s.container} contentContainerStyle={s.content}>
      {/* Main CTA */}
      <View style={s.ctaCard}>
        <Text style={s.ctaIcon}>📨</Text>
        <Text style={s.ctaTitle}>邀请通讯录好友</Text>
        <Text style={s.ctaDesc}>
          自动发送消息给通讯录中所有好友{'\n'}每条消息含你的专属推广链接
        </Text>

        {/* Message Preview */}
        <View style={s.preview}>
          <Text style={s.previewLabel}>📱 消息预览（系统自动生成）</Text>
          <View style={s.previewBubble}>
            <Text style={s.previewText}>
              "{nickname} 邀请你使用 QuantumClaw！{'\n'}
              注册即送 50000 Token，一个 Key 调用全部 AI 大模型。{'\n'}
              👉 https://t.xxx/r/<Text style={s.affCode}>{affCode}</Text>"
            </Text>
          </View>
          <View style={s.tags}>
            <Text style={s.tag}>👤 自动填你的昵称</Text>
            <Text style={s.tag}>🔗 自动生成邀请链接</Text>
            <Text style={s.tag}>💰 好友注册你得返佣</Text>
          </View>
        </View>

        {/* Start / Continue button */}
        {!progress && (
          <TouchableOpacity
            style={s.startButton}
            onPress={handleStart}
            disabled={loading}
            activeOpacity={0.8}
          >
            {loading ? (
              <ActivityIndicator color="#fff" />
            ) : (
              <Text style={s.startButtonText}>
                {lastJob ? '📨 继续未完成的任务' : '📄 阅读协议并开始发送'}
              </Text>
            )}
          </TouchableOpacity>
        )}
        <Text style={s.costHint}>
          短信费用由你的运营商收取 · 平台不收费
        </Text>
      </View>

      {/* Live Progress */}
      {progress && progress.status !== 'completed' && progress.status !== 'cancelled' && (
        <View style={s.progressCard}>
          <View style={s.progressHeader}>
            <Text style={s.progressTitle}>
              {progress.status === 'paused' ? '⏸ 已暂停' : '📤 正在发送中'}
            </Text>
            <Text style={s.progressBadge}>
              {progress.status === 'paused' ? '已暂停' : '⚡ 自动进行'}
            </Text>
          </View>
          <Text style={s.progressCount}>
            {progress.sentCount} / {progress.totalTargets}
          </Text>
          <View style={s.progressBar}>
            <View style={[s.progressFill, {
              width: `${Math.min((progress.sentCount / progress.totalTargets) * 100, 100)}%`,
            }]} />
          </View>
          <Text style={s.progressMeta}>
            批次 {progress.currentBatch}/{progress.totalBatches} · 失败 {progress.failCount} 条
          </Text>
          <View style={s.progressActions}>
            {progress.status === 'paused' ? (
              <TouchableOpacity style={s.btnResume} onPress={resumeSend}>
                <Text style={s.btnResumeText}>▶ 恢复</Text>
              </TouchableOpacity>
            ) : (
              <TouchableOpacity style={s.btnPause} onPress={pauseSend}>
                <Text style={s.btnPauseText}>⏸ 暂停</Text>
              </TouchableOpacity>
            )}
            <TouchableOpacity style={s.btnCancel} onPress={() => {
              Alert.alert('确认取消', '确定取消本次发送任务？', [
                { text: '继续发送', style: 'cancel' },
                { text: '取消任务', style: 'destructive', onPress: cancelSend },
              ])
            }}>
              <Text style={s.btnCancelText}>⏹ 取消</Text>
            </TouchableOpacity>
          </View>
        </View>
      )}

      {/* Completed / Cancelled */}
      {progress && (progress.status === 'completed' || progress.status === 'cancelled') && (
        <View style={[s.progressCard, { borderColor: progress.status === 'completed' ? '#22c55e' : '#ef4444' }]}>
          <Text style={[s.progressTitle, { color: progress.status === 'completed' ? '#16a34a' : '#dc2626' }]}>
            {progress.status === 'completed' ? '✅ 发送完成！' : '⏹ 已取消'}
          </Text>
          <Text style={s.progressMeta}>
            已发送 {progress.sentCount} / {progress.totalTargets} 位好友 · 失败 {progress.failCount} 条
          </Text>
        </View>
      )}

      {/* Stats */}
      <View style={s.statsRow}>
        <View style={s.statBox}><Text style={s.statNum}>8</Text><Text style={s.statLabel}>已邀请</Text></View>
        <View style={[s.statBox, s.statBoxActive]}><Text style={[s.statNum, { color: '#16a34a' }]}>+¥42</Text><Text style={s.statLabel}>已获得</Text></View>
        <View style={s.statBox}><Text style={[s.statNum, { color: '#2563eb' }]}>¥120</Text><Text style={s.statLabel}>待结算</Text></View>
      </View>

      {/* Reward tiers */}
      <View style={s.rewardCard}>
        <Text style={s.rewardTitle}>🎁 邀请奖励阶梯</Text>
        <View style={s.rewardGrid}>
          <View style={s.rewardItem}><Text style={s.rewardNum}>1</Text><Text style={s.rewardLabel}>邀请 1 人</Text><Text style={s.rewardVal}>+10000 Token</Text></View>
          <View style={[s.rewardItem, s.rewardItemActive]}><Text style={s.rewardNum}>5</Text><Text style={s.rewardLabel}>邀请 5 人</Text><Text style={[s.rewardVal, { color: '#d97706' }]}>+100000 Token</Text></View>
          <View style={s.rewardItem}><Text style={s.rewardNum}>20</Text><Text style={s.rewardLabel}>邀请 20 人</Text><Text style={s.rewardVal}>+¥100 现金</Text></View>
        </View>
      </View>

      {/* Last job info */}
      {lastJob && (
        <View style={s.jobInfo}>
          <Text style={s.jobInfoTitle}>📋 上次任务</Text>
          <Text style={s.jobInfoText}>任务ID: {lastJob.id}</Text>
          <Text style={s.jobInfoText}>目标: {lastJob.total_targets} 人</Text>
          <Text style={s.jobInfoText}>状态: {lastJob.status}</Text>
        </View>
      )}
    </ScrollView>
  )
}

const s = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f9fafb' },
  content: { padding: 16, paddingBottom: 32 },

  ctaCard: { backgroundColor: '#fffbeb', borderRadius: 20, padding: 20, borderWidth: 1, borderColor: '#fcd34d', marginBottom: 16, alignItems: 'center' },
  ctaIcon: { fontSize: 48, marginBottom: 8 },
  ctaTitle: { fontSize: 20, fontWeight: '700', color: '#92400e', marginBottom: 6 },
  ctaDesc: { fontSize: 13, color: '#6b7280', textAlign: 'center', marginBottom: 16 },
  preview: { backgroundColor: '#fff', borderRadius: 12, padding: 12, width: '100%', marginBottom: 16, borderWidth: 1, borderColor: '#fde68a' },
  previewLabel: { fontSize: 11, color: '#9ca3af', marginBottom: 6 },
  previewBubble: { backgroundColor: '#f9fafb', borderRadius: 8, padding: 10 },
  previewText: { fontSize: 13, color: '#374151', lineHeight: 20 },
  affCode: { color: '#d97706', fontWeight: '700' },
  tags: { flexDirection: 'row', flexWrap: 'wrap', gap: 6, marginTop: 8 },
  tag: { fontSize: 11, paddingHorizontal: 8, paddingVertical: 3, backgroundColor: '#fffbeb', borderRadius: 4, color: '#92400e' },

  startButton: { backgroundColor: '#d97706', paddingVertical: 16, borderRadius: 14, width: '100%', alignItems: 'center', shadowColor: '#d97706', shadowOffset: { width: 0, height: 4 }, shadowOpacity: 0.35, shadowRadius: 16, elevation: 8 },
  startButtonText: { color: '#fff', fontSize: 16, fontWeight: '600' },
  costHint: { fontSize: 11, color: '#9ca3af', marginTop: 8 },

  progressCard: { backgroundColor: '#fffbeb', borderRadius: 16, padding: 16, marginBottom: 16, borderWidth: 1, borderColor: '#d97706' },
  progressHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 },
  progressTitle: { fontSize: 15, fontWeight: '600', color: '#92400e' },
  progressBadge: { fontSize: 12, color: '#d97706' },
  progressCount: { fontSize: 24, fontWeight: '700', color: '#d97706', textAlign: 'center', marginBottom: 8 },
  progressBar: { height: 10, backgroundColor: '#fef3c7', borderRadius: 5, overflow: 'hidden' },
  progressFill: { height: 10, backgroundColor: '#d97706', borderRadius: 5 },
  progressMeta: { fontSize: 12, color: '#6b7280', textAlign: 'center', marginTop: 6 },
  progressActions: { flexDirection: 'row', gap: 8, marginTop: 12 },
  btnPause: { flex: 1, paddingVertical: 10, borderRadius: 10, backgroundColor: '#fff', borderWidth: 1, borderColor: '#e5e7eb', alignItems: 'center' },
  btnPauseText: { fontSize: 13, fontWeight: '500', color: '#6b7280' },
  btnResume: { flex: 1, paddingVertical: 10, borderRadius: 10, backgroundColor: '#d97706', alignItems: 'center' },
  btnResumeText: { fontSize: 13, fontWeight: '600', color: '#fff' },
  btnCancel: { flex: 1, paddingVertical: 10, borderRadius: 10, backgroundColor: '#fff', borderWidth: 1, borderColor: '#ef4444', alignItems: 'center' },
  btnCancelText: { fontSize: 13, fontWeight: '500', color: '#dc2626' },

  statsRow: { flexDirection: 'row', gap: 8, marginBottom: 16 },
  statBox: { flex: 1, backgroundColor: '#fff', borderRadius: 14, padding: 14, alignItems: 'center', borderWidth: 1, borderColor: '#f3f0ea' },
  statBoxActive: { backgroundColor: '#f0fdf4', borderColor: '#bbf7d0' },
  statNum: { fontSize: 22, fontWeight: '700', color: '#d97706' },
  statLabel: { fontSize: 12, color: '#6b7280', marginTop: 2 },

  rewardCard: { backgroundColor: '#fffbeb', borderRadius: 16, padding: 16, marginBottom: 16, borderWidth: 1, borderColor: '#fcd34d' },
  rewardTitle: { fontSize: 15, fontWeight: '600', color: '#92400e', marginBottom: 12 },
  rewardGrid: { flexDirection: 'row', gap: 8 },
  rewardItem: { flex: 1, backgroundColor: '#fff', borderRadius: 12, padding: 10, alignItems: 'center' },
  rewardItemActive: { borderWidth: 2, borderColor: '#fcd34d' },
  rewardNum: { fontSize: 20, fontWeight: '700', color: '#d97706' },
  rewardLabel: { fontSize: 11, color: '#6b7280', marginTop: 4 },
  rewardVal: { fontSize: 10, color: '#d97706', marginTop: 2 },

  jobInfo: { backgroundColor: '#fff', borderRadius: 16, padding: 16 },
  jobInfoTitle: { fontSize: 14, fontWeight: '600', color: '#111827', marginBottom: 8 },
  jobInfoText: { fontSize: 12, color: '#6b7280', marginBottom: 2 },
})
