/**
 * QuantumClaw Mobile - Login Screen
 * 
 * Features:
 *  - 本机号码一键登录 (operator gateway)
 *  - Email/Password login
 *  - WeChat login
 *  - Register link
 */

import React, { useState } from 'react'
import {
  View, Text, TextInput, TouchableOpacity, StyleSheet,
  Alert, ActivityIndicator, KeyboardAvoidingView, Platform, ScrollView,
} from 'react-native'
import { useNavigation } from '@react-navigation/native'
import type { NativeStackNavigationProp } from '@react-navigation/native-stack'
import type { AuthStackParamList } from '../../navigation/AppNavigator'
import { signIn, getSelf } from '../../lib/api-extended'
import { setStoredCredentials, initApiClient } from '../../lib/api'
import { useAuthStore } from '../../stores/auth-store'

type LoginNav = NativeStackNavigationProp<AuthStackParamList, 'Login'>

export default function LoginScreen() {
  const navigation = useNavigation<LoginNav>()
  const setUser = useAuthStore((s) => s.setUser)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [oneClickLoading, setOneClickLoading] = useState(false)

  // ── Email/Password Login ──
  const handleLogin = async () => {
    if (!email.trim() || !password.trim()) {
      Alert.alert('提示', '请输入邮箱和密码')
      return
    }
    setLoading(true)
    try {
      const res = await signIn(email.trim(), password)
      if (!res.success) {
        Alert.alert('登录失败', res.message || '请检查账号密码')
        return
      }
      await initApiClient()
      // Fetch current user info
      const selfRes = await getSelf()
      if (selfRes.success && selfRes.data) {
        const d = selfRes.data
        setUser({
          id: d.id,
          username: d.username,
          display_name: d.display_name,
          email: d.email,
          role: d.role,
          status: d.status,
          quota: d.quota,
          used_quota: d.used_quota,
          request_count: d.request_count,
        })
        await setStoredCredentials(String(d.id))
      }
    } catch (e: any) {
      Alert.alert('登录失败', e.message || '网络错误')
    } finally {
      setLoading(false)
    }
  }

  // ── 一键登录 (placeholder - 集成运营商SDK) ──
  const handleOneClickLogin = async () => {
    setOneClickLoading(true)
    // TODO: 集成阿里云/极光 本机号码认证 SDK
    // 目前模拟流程
    Alert.alert(
      '本机号码一键登录',
      '即将获取本机号码并自动登录...\n（需集成运营商认证 SDK）',
    )
    setOneClickLoading(false)
  }

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView contentContainerStyle={styles.scroll} keyboardShouldPersistTaps="handled">
        {/* Logo */}
        <View style={styles.logoContainer}>
          <View style={styles.logo}>
            <Text style={styles.logoText}>QC</Text>
          </View>
          <Text style={styles.title}>QuantumClaw</Text>
          <Text style={styles.subtitle}>AI API 聚合平台</Text>
        </View>

        {/* 一键登录 */}
        <TouchableOpacity
          style={styles.oneClickButton}
          onPress={handleOneClickLogin}
          disabled={oneClickLoading}
          activeOpacity={0.8}
        >
          {oneClickLoading ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <>
              <Text style={styles.oneClickIcon}>📱</Text>
              <View>
                <Text style={styles.oneClickText}>本机号码一键登录</Text>
                <Text style={styles.oneClickHint}>运营商认证 · 3秒完成</Text>
              </View>
            </>
          )}
        </TouchableOpacity>

        {/* Divider */}
        <View style={styles.divider}>
          <View style={styles.dividerLine} />
          <Text style={styles.dividerText}>或</Text>
          <View style={styles.dividerLine} />
        </View>

        {/* Email Login */}
        <TextInput
          style={styles.input}
          placeholder="邮箱"
          placeholderTextColor="#9ca3af"
          value={email}
          onChangeText={setEmail}
          autoCapitalize="none"
          keyboardType="email-address"
        />
        <TextInput
          style={styles.input}
          placeholder="密码"
          placeholderTextColor="#9ca3af"
          value={password}
          onChangeText={setPassword}
          secureTextEntry
        />
        <TouchableOpacity
          style={[styles.loginButton, loading && styles.buttonDisabled]}
          onPress={handleLogin}
          disabled={loading}
          activeOpacity={0.8}
        >
          {loading ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <Text style={styles.loginButtonText}>登录</Text>
          )}
        </TouchableOpacity>

        {/* Register & WeChat */}
        <View style={styles.bottomActions}>
          <TouchableOpacity onPress={() => navigation.navigate('Register')}>
            <Text style={styles.link}>注册新账号</Text>
          </TouchableOpacity>
          <TouchableOpacity onPress={() => Alert.alert('微信登录', '即将跳转微信授权...')}>
            <Text style={styles.link}>💬 微信登录</Text>
          </TouchableOpacity>
        </View>

        {/* Agreement */}
        <Text style={styles.agreement}>
          登录即表示同意 服务条款 和 隐私政策
        </Text>
      </ScrollView>
    </KeyboardAvoidingView>
  )
}

// ── Styles ──
const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#fff' },
  scroll: { flexGrow: 1, justifyContent: 'center', paddingHorizontal: 32, paddingVertical: 60 },
  logoContainer: { alignItems: 'center', marginBottom: 40 },
  logo: {
    width: 72, height: 72, borderRadius: 22,
    backgroundColor: '#d97706',
    justifyContent: 'center', alignItems: 'center', marginBottom: 16,
  },
  logoText: { fontSize: 28, fontWeight: '800', color: '#fff' },
  title: { fontSize: 24, fontWeight: '700', color: '#111827' },
  subtitle: { fontSize: 14, color: '#6b7280', marginTop: 4 },

  // One-click login
  oneClickButton: {
    flexDirection: 'row', alignItems: 'center', justifyContent: 'center',
    backgroundColor: '#d97706', paddingVertical: 16, borderRadius: 14,
    marginBottom: 16, gap: 10,
    shadowColor: '#d97706', shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3, shadowRadius: 12, elevation: 6,
  },
  oneClickIcon: { fontSize: 22 },
  oneClickText: { color: '#fff', fontSize: 15, fontWeight: '600' },
  oneClickHint: { color: 'rgba(255,255,255,0.7)', fontSize: 11 },

  // Divider
  divider: { flexDirection: 'row', alignItems: 'center', marginVertical: 20, gap: 12 },
  dividerLine: { flex: 1, height: 1, backgroundColor: '#e5e7eb' },
  dividerText: { color: '#d1d5db', fontSize: 13 },

  // Inputs
  input: {
    backgroundColor: '#f9fafb', borderRadius: 12, paddingVertical: 14, paddingHorizontal: 16,
    fontSize: 15, color: '#111827', marginBottom: 12,
    borderWidth: 1, borderColor: '#e5e7eb',
  },

  // Login button
  loginButton: {
    backgroundColor: '#d97706', paddingVertical: 14, borderRadius: 14,
    alignItems: 'center', marginTop: 4,
  },
  buttonDisabled: { opacity: 0.6 },
  loginButtonText: { color: '#fff', fontSize: 16, fontWeight: '600' },

  // Bottom
  bottomActions: {
    flexDirection: 'row', justifyContent: 'space-between', marginTop: 20,
  },
  link: { color: '#d97706', fontSize: 14, fontWeight: '500' },

  // Agreement
  agreement: {
    textAlign: 'center', color: '#9ca3af', fontSize: 12, marginTop: 24,
  },
})
