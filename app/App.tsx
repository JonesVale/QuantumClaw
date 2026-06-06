/**
 * QuantumClaw Mobile - App Entry Point
 * 
 * Boot sequence:
 *  1. Initialize API client (load SecureStore credentials)
 *  2. Fetch user self info (check if session valid)
 *  3. Render AppNavigator (Auth or Role-based tabs)
 */

import React, { useEffect, useState } from 'react'
import { StatusBar } from 'expo-status-bar'
import { View, ActivityIndicator, StyleSheet } from 'react-native'
import { SafeAreaProvider } from 'react-native-safe-area-context'

import AppNavigator from './src/navigation/AppNavigator'
import { initApiClient } from './src/lib/api'
import { useAuthStore } from './src/stores/auth-store'

export default function App() {
  const [booted, setBooted] = useState(false)
  const setUser = useAuthStore((s) => s.setUser)
  const setLoaded = useAuthStore((s) => s.setLoaded)

  useEffect(() => {
    async function boot() {
      try {
        // 1. Load stored credentials
        await initApiClient()

        // 2. Try to restore session
        // getSelf() is called by LoginScreen after login.
        // For auto-login: check if SecureStore has a saved uid/token.
        // Currently we start unauthenticated — login screen handles auth.
      } catch {
        // Ignore boot errors
      } finally {
        setLoaded(true)
        setBooted(true)
      }
    }
    boot()
  }, [])

  if (!booted) {
    return (
      <View style={styles.splash}>
        <ActivityIndicator size="large" color="#d97706" />
      </View>
    )
  }

  return (
    <SafeAreaProvider>
      <StatusBar style="dark" />
      <AppNavigator />
    </SafeAreaProvider>
  )
}

const styles = StyleSheet.create({
  splash: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#ffffff',
  },
})
