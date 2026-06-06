/**
 * QuantumClaw Mobile - Main App Navigator
 * 
 * Root structure:
 *   AuthStack (login/register)
 *   ConsumerTabs (home/keys/wallet/profile + invite)
 *   ProviderTabs (overview/channels/store/profile)
 *   EnterpriseTabs (dashboard/approvals/members/settings)
 * 
 * Role switching: Profile screen changes useAuthStore.role
 */

import React from 'react'
import { NavigationContainer } from '@react-navigation/native'
import { createNativeStackNavigator } from '@react-navigation/native-stack'
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs'
import { Text, View, StyleSheet } from 'react-native'

import { useAuthStore, UserRole } from '../stores/auth-store'
import { setNavigateToLogin } from '../lib/api'

// Screen imports (placeholders for now - will implement in Phase 1.5)
import LoginScreen from '../screens/auth/LoginScreen'
import RegisterScreen from '../screens/auth/RegisterScreen'
import HomeScreen from '../screens/consumer/HomeScreen'
import KeysScreen from '../screens/consumer/KeysScreen'
import WalletScreen from '../screens/consumer/WalletScreen'
import ProfileScreen from '../screens/consumer/ProfileScreen'
import InviteScreen from '../screens/consumer/InviteScreen'

// Lazy imports for provider/enterprise (will implement later)
const ProviderOverview = () => <Placeholder title="Provider Overview" />
const ProviderChannels = () => <Placeholder title="Channels" />
const ProviderStore = () => <Placeholder title="My Store" />
const ProviderProfile = () => <Placeholder title="Provider Profile" />

const EnterpriseDashboard = () => <Placeholder title="Enterprise Dashboard" />
const EnterpriseApprovals = () => <Placeholder title="Approvals" />
const EnterpriseMembers = () => <Placeholder title="Members" />
const EnterpriseSettings = () => <Placeholder title="Settings" />

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type AuthStackParamList = {
  Login: undefined
  Register: undefined
}

// ---------------------------------------------------------------------------
// Placeholder component for unimplemented screens
// ---------------------------------------------------------------------------

function Placeholder({ title }: { title: string }) {
  return (
    <View style={styles.center}>
      <Text style={styles.placeholder}>{title}</Text>
      <Text style={styles.subtext}>Coming soon</Text>
    </View>
  )
}

// ---------------------------------------------------------------------------
// Auth Stack
// ---------------------------------------------------------------------------

const Auth = createNativeStackNavigator<AuthStackParamList>()

function AuthStack() {
  return (
    <Auth.Navigator screenOptions={{ headerShown: false }}>
      <Auth.Screen name="Login" component={LoginScreen} />
      <Auth.Screen name="Register" component={RegisterScreen} />
    </Auth.Navigator>
  )
}

// ---------------------------------------------------------------------------
// Consumer Tabs
// ---------------------------------------------------------------------------

const ConsumerTab = createBottomTabNavigator()

function ConsumerTabs() {
  return (
    <ConsumerTab.Navigator
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: '#d97706',
        tabBarInactiveTintColor: '#9ca3af',
        tabBarStyle: {
          backgroundColor: '#ffffff',
          borderTopColor: '#f3f0ea',
          height: 56,
          paddingBottom: 6,
          paddingTop: 4,
        },
        tabBarLabelStyle: {
          fontSize: 11,
          fontWeight: '500',
        },
      }}
    >
      <ConsumerTab.Screen
        name="Home"
        component={HomeScreen}
        options={{
          tabBarLabel: '首页',
          tabBarIcon: ({ color }) => <Text style={{ fontSize: 20, color }}>🏠</Text>,
        }}
      />
      <ConsumerTab.Screen
        name="Keys"
        component={KeysScreen}
        options={{
          tabBarLabel: 'Key',
          tabBarIcon: ({ color }) => <Text style={{ fontSize: 20, color }}>🔑</Text>,
        }}
      />
      <ConsumerTab.Screen
        name="Wallet"
        component={WalletScreen}
        options={{
          tabBarLabel: '钱包',
          tabBarIcon: ({ color }) => <Text style={{ fontSize: 20, color }}>💰</Text>,
        }}
      />
      <ConsumerTab.Screen
        name="Invite"
        component={InviteScreen}
        options={{
          tabBarLabel: '邀请',
          tabBarIcon: ({ color }) => <Text style={{ fontSize: 20, color }}>📨</Text>,
        }}
      />
      <ConsumerTab.Screen
        name="Profile"
        component={ProfileScreen}
        options={{
          tabBarLabel: '我的',
          tabBarIcon: ({ color }) => <Text style={{ fontSize: 20, color }}>👤</Text>,
        }}
      />
    </ConsumerTab.Navigator>
  )
}

// ---------------------------------------------------------------------------
// Provider Tabs
// ---------------------------------------------------------------------------

const ProviderTab = createBottomTabNavigator()

function ProviderTabs() {
  return (
    <ProviderTab.Navigator
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: '#2563eb',
        tabBarInactiveTintColor: '#9ca3af',
        tabBarStyle: { backgroundColor: '#fff', height: 56, paddingBottom: 6, paddingTop: 4 },
        tabBarLabelStyle: { fontSize: 11, fontWeight: '500' },
      }}
    >
      <ProviderTab.Screen name="Overview" component={ProviderOverview}
        options={{ tabBarLabel: '概览', tabBarIcon: ({ color }) => <Text style={{fontSize:20,color}}>📊</Text> }} />
      <ProviderTab.Screen name="Channels" component={ProviderChannels}
        options={{ tabBarLabel: '渠道', tabBarIcon: ({ color }) => <Text style={{fontSize:20,color}}>🔌</Text> }} />
      <ProviderTab.Screen name="Store" component={ProviderStore}
        options={{ tabBarLabel: '店铺', tabBarIcon: ({ color }) => <Text style={{fontSize:20,color}}>🏪</Text> }} />
      <ProviderTab.Screen name="ProviderProfile" component={ProviderProfile}
        options={{ tabBarLabel: '我的', tabBarIcon: ({ color }) => <Text style={{fontSize:20,color}}>👤</Text> }} />
    </ProviderTab.Navigator>
  )
}

// ---------------------------------------------------------------------------
// Enterprise Tabs
// ---------------------------------------------------------------------------

const EnterpriseTab = createBottomTabNavigator()

function EnterpriseTabs() {
  return (
    <EnterpriseTab.Navigator
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: '#059669',
        tabBarInactiveTintColor: '#9ca3af',
        tabBarStyle: { backgroundColor: '#fff', height: 56, paddingBottom: 6, paddingTop: 4 },
        tabBarLabelStyle: { fontSize: 11, fontWeight: '500' },
      }}
    >
      <EnterpriseTab.Screen name="Dashboard" component={EnterpriseDashboard}
        options={{ tabBarLabel: '总览', tabBarIcon: ({ color }) => <Text style={{fontSize:20,color}}>📋</Text> }} />
      <EnterpriseTab.Screen name="Approvals" component={EnterpriseApprovals}
        options={{ tabBarLabel: '审批', tabBarIcon: ({ color }) => <Text style={{fontSize:20,color}}>✅</Text> }} />
      <EnterpriseTab.Screen name="Members" component={EnterpriseMembers}
        options={{ tabBarLabel: '成员', tabBarIcon: ({ color }) => <Text style={{fontSize:20,color}}>👥</Text> }} />
      <EnterpriseTab.Screen name="EntSettings" component={EnterpriseSettings}
        options={{ tabBarLabel: '设置', tabBarIcon: ({ color }) => <Text style={{fontSize:20,color}}>⚙️</Text> }} />
    </EnterpriseTab.Navigator>
  )
}

// ---------------------------------------------------------------------------
// Root Navigator
// ---------------------------------------------------------------------------

export default function AppNavigator() {
  const user = useAuthStore((s) => s.user)
  const role = useAuthStore((s) => s.role)

  // Provide the navigation reset callback for 401 handling
  React.useEffect(() => {
    setNavigateToLogin(() => {
      useAuthStore.getState().reset()
    })
  }, [])

  // Not logged in → show auth stack
  if (!user) {
    return (
      <NavigationContainer>
        <AuthStack />
      </NavigationContainer>
    )
  }

  // Logged in → show role-based tabs
  return (
    <NavigationContainer>
      {role === 'consumer' && <ConsumerTabs />}
      {role === 'provider' && <ProviderTabs />}
      {role === 'enterprise' && <EnterpriseTabs />}
    </NavigationContainer>
  )
}

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

const styles = StyleSheet.create({
  center: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#f9fafb',
  },
  placeholder: {
    fontSize: 20,
    fontWeight: '600',
    color: '#374151',
  },
  subtext: {
    fontSize: 14,
    color: '#9ca3af',
    marginTop: 4,
  },
})
