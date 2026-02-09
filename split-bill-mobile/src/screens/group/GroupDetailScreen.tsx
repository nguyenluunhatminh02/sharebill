import React, {useEffect} from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  StyleSheet,
  SafeAreaView,
  RefreshControl,
  Share,
} from 'react-native';
import {useNavigation, useRoute, RouteProp} from '@react-navigation/native';
import {NativeStackNavigationProp} from '@react-navigation/native-stack';
import Icon from 'react-native-vector-icons/Ionicons';
import {colors, spacing, fontSize, borderRadius} from '../../theme';
import {useGroupStore} from '../../store/useGroupStore';
import {useBillStore} from '../../store/useBillStore';
import {RootStackParamList} from '../../navigation/AppNavigator';

type NavigationProp = NativeStackNavigationProp<RootStackParamList>;
type RouteProps = RouteProp<RootStackParamList, 'GroupDetail'>;

export default function GroupDetailScreen() {
  const navigation = useNavigation<NavigationProp>();
  const route = useRoute<RouteProps>();
  const {groupId} = route.params;

  const {currentGroup, fetchGroup, isLoading} = useGroupStore();
  const {bills, fetchBills, balances, fetchBalances} = useBillStore();

  useEffect(() => {
    fetchGroup(groupId);
    fetchBills(groupId);
    fetchBalances(groupId);
  }, [groupId]);

  const handleShare = async () => {
    if (!currentGroup) return;
    try {
      await Share.share({
        message: `Tham gia nhóm "${currentGroup.name}" trên Split Bill!\nMã mời: ${currentGroup.invite_code}`,
      });
    } catch (error) {}
  };

  const refresh = () => {
    fetchGroup(groupId);
    fetchBills(groupId);
    fetchBalances(groupId);
  };

  const totalSpent = bills.reduce((sum, b) => sum + b.total_amount, 0);

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView
        refreshControl={
          <RefreshControl refreshing={isLoading} onRefresh={refresh} />
        }>
        {/* Group Info Card */}
        <View style={styles.infoCard}>
          <View style={styles.infoRow}>
            <View style={styles.statBox}>
              <Text style={styles.statValue}>{currentGroup?.members?.length || 0}</Text>
              <Text style={styles.statLabel}>Thành viên</Text>
            </View>
            <View style={styles.statBox}>
              <Text style={styles.statValue}>{bills.length}</Text>
              <Text style={styles.statLabel}>Hóa đơn</Text>
            </View>
            <View style={styles.statBox}>
              <Text style={styles.statValue}>
                {(totalSpent / 1000).toFixed(0)}k
              </Text>
              <Text style={styles.statLabel}>Tổng chi</Text>
            </View>
          </View>

          {/* Invite Code */}
          <TouchableOpacity style={styles.inviteRow} onPress={handleShare}>
            <Icon name="link-outline" size={18} color={colors.primary} />
            <Text style={styles.inviteCode}>
              Mã mời: {currentGroup?.invite_code}
            </Text>
            <Icon name="share-outline" size={18} color={colors.primary} />
          </TouchableOpacity>
        </View>

        {/* Action Buttons */}
        <View style={styles.actions}>
          <TouchableOpacity
            style={styles.actionBtn}
            onPress={() =>
              navigation.navigate('AddBill', {
                groupId,
                members: currentGroup?.members || [],
              })
            }>
            <Icon name="add-circle" size={20} color={colors.textInverse} />
            <Text style={styles.actionBtnText}>Thêm Bill</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.actionBtn, styles.actionBtnOCR]}
            onPress={() =>
              navigation.navigate('ScanReceipt', {
                groupId,
                groupName: currentGroup?.name || '',
              })
            }>
            <Icon name="camera" size={20} color={colors.textInverse} />
            <Text style={styles.actionBtnText}>Quét HĐ</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.actionBtn, styles.actionBtnSecondary]}
            onPress={() =>
              navigation.navigate('Balances', {
                groupId,
                groupName: currentGroup?.name || '',
              })
            }>
            <Icon name="wallet" size={20} color={colors.primary} />
            <Text style={[styles.actionBtnText, styles.actionBtnTextSecondary]}>
              Số Dư
            </Text>
          </TouchableOpacity>
        </View>

        {/* Members */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Thành Viên</Text>
          {currentGroup?.members?.map(member => (
            <View key={member.user_id} style={styles.memberRow}>
              <View style={styles.memberAvatar}>
                <Text style={styles.memberAvatarText}>
                  {(member.display_name || member.nickname || '?').charAt(0).toUpperCase()}
                </Text>
              </View>
              <View style={styles.memberInfo}>
                <Text style={styles.memberName}>
                  {member.display_name || member.nickname}
                </Text>
                <Text style={styles.memberRole}>{member.role}</Text>
              </View>
            </View>
          ))}
        </View>

        {/* Recent Bills */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Hóa Đơn Gần Đây</Text>
          {bills.length === 0 ? (
            <Text style={styles.emptyText}>Chưa có hóa đơn nào</Text>
          ) : (
            bills.slice(0, 5).map(bill => (
              <TouchableOpacity
                key={bill.id}
                style={styles.billCard}
                onPress={() => navigation.navigate('BillDetail', {billId: bill.id})}>
                <View style={styles.billIcon}>
                  <Icon name="receipt-outline" size={20} color={colors.primary} />
                </View>
                <View style={styles.billInfo}>
                  <Text style={styles.billTitle}>{bill.title}</Text>
                  <Text style={styles.billMeta}>
                    {bill.split_type === 'equal' ? 'Chia đều' : 'Theo món'} •{' '}
                    {new Date(bill.created_at).toLocaleDateString('vi-VN')}
                  </Text>
                </View>
                <Text style={styles.billAmount}>
                  {bill.total_amount.toLocaleString('vi-VN')}đ
                </Text>
              </TouchableOpacity>
            ))
          )}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {flex: 1, backgroundColor: colors.background},
  infoCard: {
    backgroundColor: colors.surface,
    margin: spacing.lg,
    borderRadius: borderRadius.lg,
    padding: spacing.lg,
    elevation: 3,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 2},
    shadowOpacity: 0.1,
    shadowRadius: 8,
  },
  infoRow: {flexDirection: 'row', justifyContent: 'space-around', marginBottom: spacing.md},
  statBox: {alignItems: 'center'},
  statValue: {fontSize: fontSize.xxl, fontWeight: '700', color: colors.primary},
  statLabel: {fontSize: fontSize.sm, color: colors.textSecondary, marginTop: 2},
  inviteRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.primaryLight + '15',
    borderRadius: borderRadius.sm,
    padding: spacing.sm,
    gap: spacing.sm,
  },
  inviteCode: {fontSize: fontSize.md, fontWeight: '600', color: colors.primary},
  actions: {
    flexDirection: 'row',
    paddingHorizontal: spacing.lg,
    gap: spacing.md,
    marginBottom: spacing.md,
  },
  actionBtn: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.primary,
    borderRadius: borderRadius.sm,
    paddingVertical: spacing.md,
    gap: spacing.xs,
  },
  actionBtnOCR: {
    backgroundColor: colors.secondary,
  },
  actionBtnSecondary: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.primary,
  },
  actionBtnText: {color: colors.textInverse, fontWeight: '600', fontSize: fontSize.md},
  actionBtnTextSecondary: {color: colors.primary},
  section: {paddingHorizontal: spacing.lg, marginBottom: spacing.lg},
  sectionTitle: {
    fontSize: fontSize.xl,
    fontWeight: '700',
    color: colors.text,
    marginBottom: spacing.md,
  },
  memberRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing.sm,
    borderBottomWidth: 1,
    borderBottomColor: colors.borderLight,
  },
  memberAvatar: {
    width: 40,
    height: 40,
    borderRadius: borderRadius.full,
    backgroundColor: colors.secondaryLight + '30',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: spacing.md,
  },
  memberAvatarText: {fontSize: fontSize.md, fontWeight: '700', color: colors.secondary},
  memberInfo: {flex: 1},
  memberName: {fontSize: fontSize.md, fontWeight: '600', color: colors.text},
  memberRole: {fontSize: fontSize.xs, color: colors.textSecondary, textTransform: 'capitalize'},
  emptyText: {fontSize: fontSize.md, color: colors.textSecondary, textAlign: 'center', paddingVertical: spacing.lg},
  billCard: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderRadius: borderRadius.sm,
    padding: spacing.md,
    marginBottom: spacing.sm,
    elevation: 1,
  },
  billIcon: {
    width: 40,
    height: 40,
    borderRadius: borderRadius.sm,
    backgroundColor: colors.primaryLight + '20',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: spacing.md,
  },
  billInfo: {flex: 1},
  billTitle: {fontSize: fontSize.md, fontWeight: '600', color: colors.text},
  billMeta: {fontSize: fontSize.xs, color: colors.textSecondary, marginTop: 2},
  billAmount: {fontSize: fontSize.lg, fontWeight: '700', color: colors.primary},
});
