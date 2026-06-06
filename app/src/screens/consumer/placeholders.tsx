/* Placeholder screens for consumer */
import React from 'react'; import {View,Text,StyleSheet} from 'react-native'
function make(name:string,icon:string){
  const C=()=>(
    <View style={s.c}><Text style={s.g}>{icon} {name}</Text><Text style={s.h}>Phase 2</Text></View>
  );return C
}
const s=StyleSheet.create({c:{flex:1,backgroundColor:'#f9fafb',justifyContent:'center',alignItems:'center',padding:20},g:{fontSize:24,fontWeight:'700',color:'#111827',marginBottom:8},h:{color:'#9ca3af',fontSize:14}})
export const KeysScreen=make('API Keys','🔑')
export const WalletScreen=make('Wallet','💰')
export const ProfileScreen=make('Profile','👤')
