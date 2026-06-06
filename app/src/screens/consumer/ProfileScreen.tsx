/**
 * Consumer ProfileScreen - Settings & Role Switch
 */

import React from 'react'
import {
  View, Text, StyleSheet, ScrollView, TouchableOpacity, Alert,
} from 'react-native'
import { useAuthStore, UserRole } from '../../stores/auth-store'
import { clearStoredCredentials } from '../../lib/api'
import { signOut } from '../../lib/api-extended'

export default function ProfileScreen() {
  const user = useAuthStore((s) => s.user)
  const role = useAuthStore((s) => s.role)
  const setRole = useAuthStore((s) => s.setRole)
  const reset = useAuthStore((s) => s.reset)

  const handleRoleSwitch = (newRole: UserRole) => {
    setRole(newRole)
    Alert.alert('角色已切换', `当前: ${newRole === 'consumer' ? '个人用户' : newRole === 'provider' ? '渠道商' : '企业'}`)
  }

  const handleLogout = () => {
    Alert.alert('确认退出', '确定要退出登录吗？', [
      { text: '取消', style: 'cancel' },
      { text: '退出', style: 'destructive', onPress: async () => {
        try { await signOut() } catch { /* ignore */ }
        await clearStoredCredentials()
        reset()
      }},
    ])
  }

  const nickname = user?.display_name || user?.username || '用户'
  const availableRoles: { role: UserRole; label: string; icon: string }[] = [
    { role: 'consumer', label: '个人用户', icon: '👤' },
  ]
  // Provider/Enterprise only show if user has those roles
  // For now, always show as options (backend determines real access)
  availableRoles.push({ role: 'provider', label: '渠道商', icon: '🏪' })
  availableRoles.push({ role: 'enterprise', label: '企业', icon: '🏢' })

  return (
    <ScrollView style={s.container} contentContainerStyle={s.content}>
      {/* Profile Header */}
      <View style={s.profileHeader}>
        <View style={s.avatar}>
          <Text style={s.avatarText}>{nickname[0]}</Text>
        </View>
        <Text style={s.displayName}>{nickname}</Text>
        <Text style={s.email}>{user?.email || '未设置邮箱'}</Text>
      </View>

      {/* Role Switcher */}
      <View style={s.section}>
        <Text style={s.sectionTitle}>🔄 角色切换</Text>
        <Text style={s.sectionHint}>当前: {role === 'consumer' ? '个人用户' : role === 'provider' ? '渠道商' : '企业'}</Text>
        <View style={s.roleGrid}>
          {availableRoles.map((r) => (
            <TouchableOpacity
              key={r.role}
              style={[s.roleBtn, role === r.role && s.roleBtnActive]}
              onPress={() => handleRoleSwitch(r.role)}
            >
              <Text style={s.roleIcon}>{r.icon}</Text>
              <Text style={[s.roleLabel, role === r.role && s.roleLabelActive]}>{r.label}</Text>
            </TouchableOpacity>
          ))}
        </View>
      </View>

      {/* Account Info */}
      <View style={s.section}>
        <Text style={s.sectionTitle}>📋 账户信息</Text>
        <View style={s.infoRow}><Text style={s.infoLabel}>用户ID</Text><Text style={s.infoValue}>{user?.id}</Text></View>
        <View style={s.infoRow}><Text style={s.infoLabel}>用户名</Text><Text style={s.infoValue}>{user?.username}</Text></View>
        <View style={s.infoRow}><Text style={s.infoLabel}>角色</Text><Text style={s.infoValue}>{user?.role === 100 ? '超级管理员' : user?.role === 10 ? '管理员' : '普通用户'}</Text></View>
        <View style={s.infoRow}><Text style={s.infoLabel}>状态</Text><Text style={[s.infoValue, { color: user?.status === 1 ? '#16a34a' : '#dc2626' }]}>{user?.status === 1 ? '正常' : '禁用'}</Text></View>
      </View>

      {/* Logout */}
      <TouchableOpacity style={s.logoutBtn} onPress={handleLogout}>
        <Text style={s.logoutText}>🚪 退出登录</Text>
      </TouchableOpacity>
    </ScrollView>
  )
}

const s = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#f9fafb' },
  content: { padding: 16, paddingBottom: 32 },
  profileHeader: { alignItems: 'center', paddingVertical: 24, marginBottom: 16 },
  avatar: { width: 64, height: 64, borderRadius: 32, backgroundColor: '#d97706', justifyContent: 'center', alignItems: 'center', marginBottom: 12 },
  avatarText: { fontSize: 24, fontWeight: '700', color: '#fff' },
  displayName: { fontSize: 20, fontWeight: '700', color: '#111827' },
  email: { fontSize: 13, color: '#9ca3af', marginTop: 4 },

  section: { backgroundColor: '#fff', borderRadius: 16, padding: 16, marginBottom: 16 },
  sectionTitle: { fontSize: 15, fontWeight: '600', color: '#111827', marginBottom: 4 },
  sectionHint: { fontSize: 12, color: '#9ca3af', marginBottom: 12 },

  roleGrid: { flexDirection: 'row', gap: 8 },
  roleBtn: { flex: 1, padding: 14, borderRadius: 14, backgroundColor: '#f9fafb', alignItems: 'center', borderWidth: 1, borderColor: '#e5e7eb' },
  roleBtnActive: { backgroundColor: '#fffbeb', borderColor: '#fcd34d' },
  roleIcon: { fontSize: 28, marginBottom: 6 },
  roleLabel: { fontSize: 13, fontWeight: '500', color: '#6b7280' },
  roleLabelActive: { color: '#d97706', fontWeight: '600' },

  infoRow: { flexDirection: 'row', justifyContent: 'space-between', paddingVertical: 10, borderBottomWidth: 1, borderBottomColor: '#f3f0ea' },
  infoLabel: { fontSize: 14, color: '#6b7280' },
  infoValue: { fontSize: 14, color: '#111827', fontWeight: '500' },

  logoutBtn: { padding: 16, borderRadius: 14, backgroundColor: '#fef2f2', alignItems: 'center', borderWidth: 1, borderColor: '#fecaca' },
  logoutText: { fontSize: 15, fontWeight: '600', color: '#dc2626' },
})
