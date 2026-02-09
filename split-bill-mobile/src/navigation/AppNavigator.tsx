import React from 'react';
import {NavigationContainer} from '@react-navigation/native';
import {createNativeStackNavigator} from '@react-navigation/native-stack';
import {createBottomTabNavigator} from '@react-navigation/bottom-tabs';
import Icon from 'react-native-vector-icons/Ionicons';

import {useAuthStore} from '../store/useAuthStore';
import {colors} from '../theme';
import {OCRResult} from '../types';

// Screens
import LoginScreen from '../screens/auth/LoginScreen';
import HomeScreen from '../screens/home/HomeScreen';
import GroupListScreen from '../screens/group/GroupListScreen';
import GroupDetailScreen from '../screens/group/GroupDetailScreen';
import CreateGroupScreen from '../screens/group/CreateGroupScreen';
import AddBillScreen from '../screens/bill/AddBillScreen';
import BillDetailScreen from '../screens/bill/BillDetailScreen';
import BalancesScreen from '../screens/settlement/BalancesScreen';
import ProfileScreen from '../screens/profile/ProfileScreen';
import ScanReceiptScreen from '../screens/ocr/ScanReceiptScreen';
import ReviewOCRScreen from '../screens/ocr/ReviewOCRScreen';
import PaymentScreen from '../screens/payment/PaymentScreen';
import ActivityScreen from '../screens/activity/ActivityScreen';
import StatisticsScreen from '../screens/stats/StatisticsScreen';

// Types
export type RootStackParamList = {
  Auth: undefined;
  Main: undefined;
  GroupDetail: {groupId: string; groupName: string};
  CreateGroup: undefined;
  AddBill: {groupId: string; members: any[]};
  BillDetail: {billId: string};
  Balances: {groupId: string; groupName: string};
  ScanReceipt: {groupId: string; groupName: string};
  ReviewOCR: {ocrResult: OCRResult; groupId: string; groupName: string};
  Payment: {
    toUserId: string;
    toUserName: string;
    amount: number;
    groupId: string;
    groupName: string;
  };
  Activity: {groupId?: string; groupName?: string};
  Statistics: {groupId: string; groupName: string};
};

export type MainTabParamList = {
  Home: undefined;
  Groups: undefined;
  Profile: undefined;
};

const Stack = createNativeStackNavigator<RootStackParamList>();
const Tab = createBottomTabNavigator<MainTabParamList>();

function MainTabs() {
  return (
    <Tab.Navigator
      screenOptions={({route}) => ({
        tabBarIcon: ({focused, color, size}) => {
          let iconName = 'home';
          if (route.name === 'Home') {
            iconName = focused ? 'home' : 'home-outline';
          } else if (route.name === 'Groups') {
            iconName = focused ? 'people' : 'people-outline';
          } else if (route.name === 'Profile') {
            iconName = focused ? 'person' : 'person-outline';
          }
          return <Icon name={iconName} size={size} color={color} />;
        },
        tabBarActiveTintColor: colors.primary,
        tabBarInactiveTintColor: colors.textLight,
        headerShown: false,
        tabBarStyle: {
          borderTopWidth: 0,
          elevation: 10,
          shadowOpacity: 0.1,
          shadowRadius: 10,
          height: 60,
          paddingBottom: 8,
        },
      })}>
      <Tab.Screen name="Home" component={HomeScreen} />
      <Tab.Screen name="Groups" component={GroupListScreen} />
      <Tab.Screen name="Profile" component={ProfileScreen} />
    </Tab.Navigator>
  );
}

export default function AppNavigator() {
  const isAuthenticated = useAuthStore(state => state.isAuthenticated);

  return (
    <NavigationContainer>
      <Stack.Navigator
        screenOptions={{
          headerStyle: {backgroundColor: colors.primary},
          headerTintColor: colors.textInverse,
          headerTitleStyle: {fontWeight: '600'},
        }}>
        {!isAuthenticated ? (
          <Stack.Screen
            name="Auth"
            component={LoginScreen}
            options={{headerShown: false}}
          />
        ) : (
          <>
            <Stack.Screen
              name="Main"
              component={MainTabs}
              options={{headerShown: false}}
            />
            <Stack.Screen
              name="GroupDetail"
              component={GroupDetailScreen}
              options={({route}) => ({title: route.params.groupName})}
            />
            <Stack.Screen
              name="CreateGroup"
              component={CreateGroupScreen}
              options={{title: 'Tạo Nhóm Mới'}}
            />
            <Stack.Screen
              name="AddBill"
              component={AddBillScreen}
              options={{title: 'Thêm Hóa Đơn'}}
            />
            <Stack.Screen
              name="BillDetail"
              component={BillDetailScreen}
              options={{title: 'Chi Tiết Hóa Đơn'}}
            />
            <Stack.Screen
              name="Balances"
              component={BalancesScreen}
              options={({route}) => ({title: `${route.params.groupName} - Số Dư`})}
            />
            <Stack.Screen
              name="ScanReceipt"
              component={ScanReceiptScreen}
              options={{title: '📷 Quét Hóa Đơn'}}
            />
            <Stack.Screen
              name="ReviewOCR"
              component={ReviewOCRScreen}
              options={{title: '✅ Xác Nhận Hóa Đơn'}}
            />
            <Stack.Screen
              name="Payment"
              component={PaymentScreen}
              options={{title: '💰 Thanh Toán'}}
            />
            <Stack.Screen
              name="Activity"
              component={ActivityScreen}
              options={({route}) => ({
                title: route.params.groupName
                  ? `📋 ${route.params.groupName}`
                  : '📋 Hoạt Động',
              })}
            />
            <Stack.Screen
              name="Statistics"
              component={StatisticsScreen}
              options={({route}) => ({
                title: `📊 ${route.params.groupName}`,
              })}
            />
          </>
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
}
