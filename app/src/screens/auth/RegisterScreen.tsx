/**
 * Register Screen - placeholder
 */
import React from 'react'
import { View, Text, TouchableOpacity, StyleSheet, TextInput } from 'react-native'
import { useNavigation } from '@react-navigation/native'
import type { NativeStackNavigationProp } from '@react-navigation/native-stack'
import type { AuthStackParamList } from '../../navigation/AppNavigator'

type Nav = NativeStackNavigationProp<AuthStackParamList, 'Register'>

export default function RegisterScreen() {
  const nav = useNavigation<Nav>()
  return (
    <View style={styles.container}>
      <Text style={styles.title}>创建账号</Text>
      <TextInput style={styles.input} placeholder="用户名" placeholderTextColor="#9ca3af" />
      <TextInput style={styles.input} placeholder="邮箱" placeholderTextColor="#9ca3af" keyboardType="email-address" />
      <TextInput style={styles.input} placeholder="密码" placeholderTextColor="#9ca3af" secureTextEntry />
      <TouchableOpacity style={styles.button} onPress={() => nav.goBack()}>
        <Text style={styles.buttonText}>注册</Text>
      </TouchableOpacity>
      <TouchableOpacity onPress={() => nav.goBack()} style={styles.back}>
        <Text style={styles.backText}>已有账号？返回登录</Text>
      </TouchableOpacity>
    </View>
  )
}
const styles = StyleSheet.create({
  container: { flex:1, backgroundColor:'#fff', paddingHorizontal:32, paddingTop:80 },
  title: { fontSize:24, fontWeight:'700', marginBottom:24, textAlign:'center' },
  input: { backgroundColor:'#f9fafb', borderRadius:12, paddingVertical:14, paddingHorizontal:16, fontSize:15, marginBottom:12, borderWidth:1, borderColor:'#e5e7eb' },
  button: { backgroundColor:'#d97706', paddingVertical:14, borderRadius:14, alignItems:'center', marginTop:8 },
  buttonText: { color:'#fff', fontSize:16, fontWeight:'600' },
  back: { marginTop:16, alignItems:'center' },
  backText: { color:'#d97706', fontSize:14 },
})
