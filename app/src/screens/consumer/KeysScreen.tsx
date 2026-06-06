/**
 * Consumer KeysScreen - API Key Management
 *
 * Full CRUD: list, create, delete API Keys
 * Data source: getTokens(), createToken(), deleteToken()
 */

import React, { useEffect, useState, useCallback } from 'react'
import {
  View, Text, StyleSheet, ScrollView, TouchableOpacity, RefreshControl,
  Alert, ActivityIndicator, TextInput, Modal,
} from 'react-native'
import { getTokens, createToken, deleteToken, Token } from '../../lib/api-extended'

export default function KeysScreen() {
  const [keys, setKeys] = useState<Token[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')
  const [creating, setCreating] = useState(false)

  const fetchKeys = useCallback(async () => {
    try {
      const res = await getTokens()
      if (res.success && res.data) setKeys(res.data)
    } catch { /* ignore */ }
    finally { setLoading(false); setRefreshing(false) }
  }, [])

  useEffect(() => { fetchKeys() }, [])

  const handleCreate = async () => {
    if (!newName.trim()) { Alert.alert('提示', '请输入Key名称'); return }
    setCreating(true)
    try {
      const res = await createToken({ name: newName.trim() })
      if (res.success) {
        Alert.alert('成功', 'API Key 已创建')
        setShowCreate(false); setNewName('')
        fetchKeys()
      } else {
        Alert.alert('创建失败', res.message || '请重试')
      }
    } catch { Alert.alert('创建失败', '网络错误') }
    finally { setCreating(false) }
  }

  const handleDelete = (id: number, name: string) => {
    Alert.alert('确认删除', `删除 Key "${name}"？此操作不可撤销。`, [
      { text: '取消', style: 'cancel' },
      { text: '删除', style: 'destructive', onPress: async () => {
        try {
          const res = await deleteToken(id)
          if (res.success) {
            setKeys((prev) => prev.filter((k) => k.id !== id))
          } else {
            Alert.alert('删除失败', res.message)
          }
        } catch { Alert.alert('删除失败', '网络错误') }
      }},
    ])
  }

  const handleCopy = (key: string) => {
    // In real app: Clipboard.setStringAsync(key)
    Alert.alert('复制', `Key: ${key.substring(0, 8)}****`)
  }

  if (loading) {
    return <View style={s.center}><ActivityIndicator size="large" color="#d97706" /></View>
  }

  return (
    <ScrollView
      style={s.container}
      contentContainerStyle={s.content}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={() => { setRefreshing(true); fetchKeys() }} tintColor="#d97706" />}
    >
      {/* Header */}
      <View style={s.header}>
        <Text style={s.title}>🔑 API Keys</Text>
        <TouchableOpacity style={s.addBtn} onPress={() => setShowCreate(true)}>
          <Text style={s.addBtnText}>+ 新建</Text>
        </TouchableOpacity>
      </View>

      {/* Key List */}
      {keys.length === 0 && (
        <View style={s.emptyState}>
          <Text style={s.emptyIcon}>🔑</Text>
          <Text style={s.emptyText}>暂无 API Key</Text>
          <Text style={s.emptyHint}>点击右上角"新建"创建第一个 Key</Text>
        </View>
      )}
      {keys.map((k) => (
        <View key={k.id} style={s.keyCard}>
          <View style={s.keyHeader}>
            <Text style={s.keyName}>{k.name || '未命名'}</Text>
            <View style={[s.statusDot, { backgroundColor: k.status === 1 ? '#22c55e' : '#ef4444' }]} />
          </View>
          <Text style={s.keyValue}>
            {k.key ? k.key.substring(0, 12) + '****' : '***'}
          </Text>
          <View style={s.keyMeta}>
            <Text style={s.keyMetaText}>
              剩余: {k.unlimited_quota ? '∞' : k.remain_quota != null ? `${(k.remain_quota / 1000000).toFixed(1)}M` : '—'}
            </Text>
            <Text style={s.keyMetaText}>
              {k.expired_time ? `过期: ${new Date(k.expired_time * 1000).toLocaleDateString()}` : '永不过期'}
            </Text>
          </View>
          <View style={s.keyActions}>
            <TouchableOpacity style={s.actionCopy} onPress={() => handleCopy(k.key)}>
              <Text style={s.actionCopyText}>📋 复制</Text>
            </TouchableOpacity>
            <TouchableOpacity style={s.actionDelete} onPress={() => handleDelete(k.id, k.name)}>
              <Text style={s.actionDeleteText}>🗑 删除</Text>
            </TouchableOpacity>
          </View>
        </View>
      ))}

      {/* Create Modal */}
      <Modal visible={showCreate} transparent animationType="fade">
        <View style={s.modalOverlay}>
          <View style={s.modal}>
            <Text style={s.modalTitle}>🔑 新建 API Key</Text>
            <TextInput
              style={s.modalInput}
              placeholder="Key 名称"
              placeholderTextColor="#9ca3af"
              value={newName}
              onChangeText={setNewName}
              autoFocus
            />
            <View style={s.modalActions}>
              <TouchableOpacity style={s.modalCancel} onPress={() => { setShowCreate(false); setNewName('') }}>
                <Text style={s.modalCancelText}>取消</Text>
              </TouchableOpacity>
              <TouchableOpacity style={s.modalConfirm} onPress={handleCreate} disabled={creating}>
                {creating ? <ActivityIndicator color="#fff" /> : <Text style={s.modalConfirmText}>创建</Text>}
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
    </ScrollView>
  )
}

const s = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f9fafb' },
  content: { padding: 16, paddingBottom: 32 },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: '#f9fafb' },

  header: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 },
  title: { fontSize: 20, fontWeight: '700', color: '#111827' },
  addBtn: { backgroundColor: '#d97706', paddingHorizontal: 16, paddingVertical: 8, borderRadius: 10 },
  addBtnText: { color: '#fff', fontSize: 14, fontWeight: '600' },

  emptyState: { alignItems: 'center', paddingVertical: 60 },
  emptyIcon: { fontSize: 48, marginBottom: 12 },
  emptyText: { fontSize: 16, fontWeight: '600', color: '#6b7280' },
  emptyHint: { fontSize: 13, color: '#9ca3af', marginTop: 4 },

  keyCard: { backgroundColor: '#fff', borderRadius: 16, padding: 16, marginBottom: 12 },
  keyHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 },
  keyName: { fontSize: 15, fontWeight: '600', color: '#111827' },
  statusDot: { width: 8, height: 8, borderRadius: 4 },
  keyValue: { fontSize: 13, color: '#6b7280', fontFamily: 'monospace', marginBottom: 8 },
  keyMeta: { flexDirection: 'row', gap: 16, marginBottom: 12 },
  keyMetaText: { fontSize: 12, color: '#9ca3af' },

  keyActions: { flexDirection: 'row', gap: 8, borderTopWidth: 1, borderTopColor: '#f3f0ea', paddingTop: 12 },
  actionCopy: { flex: 1, paddingVertical: 8, borderRadius: 8, backgroundColor: '#fffbeb', alignItems: 'center' },
  actionCopyText: { fontSize: 13, color: '#92400e', fontWeight: '500' },
  actionDelete: { flex: 1, paddingVertical: 8, borderRadius: 8, backgroundColor: '#fef2f2', alignItems: 'center' },
  actionDeleteText: { fontSize: 13, color: '#dc2626', fontWeight: '500' },

  modalOverlay: { flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: 'rgba(0,0,0,0.5)', padding: 32 },
  modal: { backgroundColor: '#fff', borderRadius: 20, padding: 24, width: '100%', maxWidth: 340 },
  modalTitle: { fontSize: 18, fontWeight: '600', marginBottom: 16 },
  modalInput: { backgroundColor: '#f9fafb', borderRadius: 12, padding: 14, fontSize: 15, borderWidth: 1, borderColor: '#e5e7eb', marginBottom: 16 },
  modalActions: { flexDirection: 'row', gap: 8 },
  modalCancel: { flex: 1, padding: 12, borderRadius: 12, borderWidth: 1, borderColor: '#e5e7eb', alignItems: 'center' },
  modalCancelText: { fontSize: 14, fontWeight: '500', color: '#6b7280' },
  modalConfirm: { flex: 1, padding: 12, borderRadius: 12, backgroundColor: '#d97706', alignItems: 'center' },
  modalConfirmText: { fontSize: 14, fontWeight: '600', color: '#fff' },
})
