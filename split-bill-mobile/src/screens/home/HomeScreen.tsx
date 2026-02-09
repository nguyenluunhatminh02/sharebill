import React, {useEffect, useState, useCallback} from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  StyleSheet,
  SafeAreaView,
  RefreshControl,
  ActivityIndicator,
} from 'react-native';
import {useNavigation} from '@react-navigation/native';
import {NativeStackNavigationProp} from '@react-navigation/native-stack';
import Icon from 'react-native-vector-icons/Ionicons';
import {colors, spacing, fontSize, borderRadius} from '../../theme';
import {useAuthStore} from '../../store/useAuthStore';
import {useGroupStore} from '../../store/useGroupStore';
import {RootStackParamList} from '../../navigation/AppNavigator';
import {activityAPI, statsAPI} from '../../api/services';
import {Activity, UserOverallStats} from '../../types';

type NavigationProp = NativeStackNavigationProp<RootStackParamList>;

const formatVND = (amount: number): string => {
  if (amount >= 1_000_000) {
    return (amount / 1_000_000).toFixed(1).replace(/\.0$/, '') + 'tr';
  }
  if (amount >= 1_000) {
    return (amount / 1_000).toFixed(0) + 'k';
  }
  return amount.toLocaleString('vi-VN');
};

const getActivityIcon = (type: string): {name: string; color: string} => {
  switch (type) {
    case 'bill_created':
      return {name: 'receipt-outline', color: colors.primary};
    case 'bill_deleted':
      return {name: 'trash-outline', color: colors.error};
    case 'bill_updated':
      return {name: 'create-outline', color: colors.warning};
    case 'member_joined':
      return {name: 'person-add-outline', color: colors.success};
    case 'member_left':
      return {name: 'person-remove-outline', color: colors.error};
    case 'payment_sent':
      return {name: 'send-outline', color: colors.secondary};
    case 'payment_confirmed':
      return {name: 'checkmark-circle-outline', color: colors.success};
    case 'group_created':
      return {name: 'people-outline', color: colors.primary};
    case 'settlement_created':
      return {name: 'swap-horizontal-outline', color: colors.accent};
    default:
      return {name: 'information-circle-outline', color: colors.textSecondary};
  }
};

export default function HomeScreen() {
  const navigation = useNavigation<NavigationProp>();
  const {user} = useAuthStore();
  const {groups, fetchGroups, isLoading} = useGroupStore();

  const [userStats, setUserStats] = useState<UserOverallStats | null>(null);
  const [recentActivities, setRecentActivities] = useState<Activity[]>([]);
  const [statsLoading, setStatsLoading] = useState(false);

  const loadData = useCallback(async () => {
    setStatsLoading(true);
    try {
      const [statsRes, activityRes] = await Promise.allSettled([
        statsAPI.getUserStats(),
        activityAPI.getMyActivities(5),
      ]);

      if (statsRes.status === 'fulfilled' && statsRes.value.data?.data) {
        setUserStats(statsRes.value.data.data);
      }
      if (
        activityRes.status === 'fulfilled' &&
        activityRes.value.data?.data
      ) {
        setRecentActivities(activityRes.value.data.data);
      }
    } catch (_e) {
      // silently fail - stats are non-critical
    } finally {
      setStatsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchGroups();
    loadData();
  }, []);

  const onRefresh = useCallback(async () => {
    await Promise.all([fetchGroups(), loadData()]);
  }, [fetchGroups, loadData]);

  const recentGroups = groups.slice(0, 3);

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView
        refreshControl={
          <RefreshControl refreshing={isLoading} onRefresh={onRefresh} />
        }>
        {/* Header */}
        <View style={styles.header}>
          <View>
            <Text style={styles.greeting}>Xin chào 👋</Text>
            <Text style={styles.userName}>{user?.display_name || 'User'}</Text>
          </View>
          <TouchableOpacity style={styles.notifButton}>
            <Icon name="notifications-outline" size={24} color={colors.text} />
          </TouchableOpacity>
        </View>

        {/* Stats Summary Card */}
        {userStats && (
          <View style={styles.statsCard}>
            <View style={styles.statsRow}>
              <View style={styles.statItem}>
                <View
                  style={[
                    styles.statIconWrap,
                    {backgroundColor: colors.primary + '20'},
                  ]}>
                  <Icon name="wallet-outline" size={20} color={colors.primary} />
                </View>
                <Text style={styles.statValue}>
                  {formatVND(userStats.total_spent)}đ
                </Text>
                <Text style={styles.statLabel}>Đã chi</Text>
              </View>
              <View style={styles.statDivider} />
              <View style={styles.statItem}>
                <View
                  style={[
                    styles.statIconWrap,
                    {backgroundColor: colors.error + '20'},
                  ]}>
                  <Icon
                    name="trending-down-outline"
                    size={20}
                    color={colors.error}
                  />
                </View>
                <Text style={[styles.statValue, {color: colors.error}]}>
                  {formatVND(userStats.total_owed)}đ
                </Text>
                <Text style={styles.statLabel}>Đang nợ</Text>
              </View>
              <View style={styles.statDivider} />
              <View style={styles.statItem}>
                <View
                  style={[
                    styles.statIconWrap,
                    {backgroundColor: colors.secondary + '20'},
                  ]}>
                  <Icon
                    name="people-outline"
                    size={20}
                    color={colors.secondary}
                  />
                </View>
                <Text style={styles.statValue}>{userStats.total_groups}</Text>
                <Text style={styles.statLabel}>Nhóm</Text>
              </View>
            </View>
          </View>
        )}

        {statsLoading && !userStats && (
          <View style={styles.statsCardPlaceholder}>
            <ActivityIndicator size="small" color={colors.primary} />
            <Text style={styles.loadingText}>Đang tải thống kê...</Text>
          </View>
        )}

        {/* Quick Actions */}
        <View style={styles.quickActions}>
          <TouchableOpacity
            style={styles.actionCard}
            onPress={() => navigation.navigate('CreateGroup')}>
            <View style={[styles.actionIcon, {backgroundColor: colors.primary + '20'}]}>
              <Icon name="people-outline" size={24} color={colors.primary} />
            </View>
            <Text style={styles.actionText}>Tạo Nhóm</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.actionCard}
            onPress={() => {
              if (groups.length > 0) {
                navigation.navigate('ScanReceipt', {
                  groupId: groups[0].id,
                  groupName: groups[0].name,
                });
              } else {
                navigation.navigate('CreateGroup');
              }
            }}>
            <View style={[styles.actionIcon, {backgroundColor: colors.secondary + '20'}]}>
              <Icon name="scan-outline" size={24} color={colors.secondary} />
            </View>
            <Text style={styles.actionText}>Scan Bill</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.actionCard}
            onPress={() => {
              if (groups.length > 0) {
                navigation.navigate('Balances', {
                  groupId: groups[0].id,
                  groupName: groups[0].name,
                });
              }
            }}>
            <View style={[styles.actionIcon, {backgroundColor: colors.accent + '20'}]}>
              <Icon name="wallet-outline" size={24} color={colors.accent} />
            </View>
            <Text style={styles.actionText}>Nợ Của Tôi</Text>
          </TouchableOpacity>
        </View>

        {/* Recent Groups */}
        <View style={styles.section}>
          <View style={styles.sectionHeader}>
            <Text style={styles.sectionTitle}>Nhóm Gần Đây</Text>
            <TouchableOpacity onPress={() => navigation.navigate('Main')}>
              <Text style={styles.seeAll}>Xem tất cả</Text>
            </TouchableOpacity>
          </View>

          {recentGroups.length === 0 ? (
            <View style={styles.emptyState}>
              <Icon name="people-outline" size={48} color={colors.textLight} />
              <Text style={styles.emptyText}>Chưa có nhóm nào</Text>
              <TouchableOpacity
                style={styles.createButton}
                onPress={() => navigation.navigate('CreateGroup')}>
                <Text style={styles.createButtonText}>+ Tạo nhóm đầu tiên</Text>
              </TouchableOpacity>
            </View>
          ) : (
            recentGroups.map(group => (
              <TouchableOpacity
                key={group.id}
                style={styles.groupCard}
                onPress={() =>
                  navigation.navigate('GroupDetail', {
                    groupId: group.id,
                    groupName: group.name,
                  })
                }>
                <View style={styles.groupAvatar}>
                  <Text style={styles.groupAvatarText}>
                    {group.name.charAt(0).toUpperCase()}
                  </Text>
                </View>
                <View style={styles.groupInfo}>
                  <Text style={styles.groupName}>{group.name}</Text>
                  <Text style={styles.groupMembers}>
                    {group.members?.length || 0} thành viên
                  </Text>
                </View>
                <Icon name="chevron-forward" size={20} color={colors.textLight} />
              </TouchableOpacity>
            ))
          )}
        </View>

        {/* Recent Activity */}
        {recentActivities.length > 0 && (
          <View style={styles.section}>
            <View style={styles.sectionHeader}>
              <Text style={styles.sectionTitle}>Hoạt Động Gần Đây</Text>
            </View>

            {recentActivities.map(activity => {
              const iconInfo = getActivityIcon(activity.type);
              return (
                <View key={activity.id} style={styles.activityCard}>
                  <View
                    style={[
                      styles.activityIconWrap,
                      {backgroundColor: iconInfo.color + '15'},
                    ]}>
                    <Icon
                      name={iconInfo.name}
                      size={20}
                      color={iconInfo.color}
                    />
                  </View>
                  <View style={styles.activityContent}>
                    <Text style={styles.activityTitle} numberOfLines={1}>
                      {activity.title}
                    </Text>
                    <Text style={styles.activityDetail} numberOfLines={1}>
                      {activity.detail}
                    </Text>
                    <View style={styles.activityMeta}>
                      {activity.group_name && (
                        <Text style={styles.activityGroup}>
                          {activity.group_name}
                        </Text>
                      )}
                      <Text style={styles.activityTime}>
                        {activity.time_ago}
                      </Text>
                    </View>
                  </View>
                  {activity.amount != null && activity.amount > 0 && (
                    <Text style={styles.activityAmount}>
                      {formatVND(activity.amount)}đ
                    </Text>
                  )}
                </View>
              );
            })}
          </View>
        )}

        {/* Top Groups by Spending */}
        {userStats && userStats.top_groups && userStats.top_groups.length > 0 && (
          <View style={[styles.section, {marginBottom: spacing.xxl}]}>
            <View style={styles.sectionHeader}>
              <Text style={styles.sectionTitle}>Top Nhóm Chi Tiêu</Text>
            </View>

            {userStats.top_groups.slice(0, 3).map((groupInfo, index) => {
              const maxSpent = userStats.top_groups[0]?.total_spent || 1;
              const barWidth = (groupInfo.total_spent / maxSpent) * 100;
              return (
                <View key={groupInfo.group_id} style={styles.topGroupCard}>
                  <View style={styles.topGroupHeader}>
                    <View style={styles.topGroupRank}>
                      <Text style={styles.topGroupRankText}>{index + 1}</Text>
                    </View>
                    <View style={styles.topGroupInfo}>
                      <Text style={styles.topGroupName} numberOfLines={1}>
                        {groupInfo.group_name}
                      </Text>
                      <Text style={styles.topGroupBills}>
                        {groupInfo.total_bills} hóa đơn
                      </Text>
                    </View>
                    <Text style={styles.topGroupAmount}>
                      {formatVND(groupInfo.total_spent)}đ
                    </Text>
                  </View>
                  <View style={styles.topGroupBarBg}>
                    <View
                      style={[
                        styles.topGroupBarFill,
                        {
                          width: `${barWidth}%`,
                          backgroundColor:
                            index === 0
                              ? colors.primary
                              : index === 1
                              ? colors.secondary
                              : colors.accent,
                        },
                      ]}
                    />
                  </View>
                </View>
              );
            })}
          </View>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: spacing.lg,
    paddingTop: spacing.lg,
    paddingBottom: spacing.md,
  },
  greeting: {
    fontSize: fontSize.md,
    color: colors.textSecondary,
  },
  userName: {
    fontSize: fontSize.xxl,
    fontWeight: '700',
    color: colors.text,
  },
  notifButton: {
    width: 44,
    height: 44,
    borderRadius: borderRadius.full,
    backgroundColor: colors.surface,
    justifyContent: 'center',
    alignItems: 'center',
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.1,
    shadowRadius: 4,
  },

  // Stats Summary Card
  statsCard: {
    marginHorizontal: spacing.lg,
    backgroundColor: colors.surface,
    borderRadius: borderRadius.lg,
    padding: spacing.lg,
    elevation: 3,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 2},
    shadowOpacity: 0.1,
    shadowRadius: 8,
  },
  statsRow: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  statItem: {
    flex: 1,
    alignItems: 'center',
  },
  statIconWrap: {
    width: 40,
    height: 40,
    borderRadius: borderRadius.full,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: spacing.xs,
  },
  statValue: {
    fontSize: fontSize.lg,
    fontWeight: '700',
    color: colors.text,
    marginTop: 2,
  },
  statLabel: {
    fontSize: fontSize.xs,
    color: colors.textSecondary,
    marginTop: 2,
  },
  statDivider: {
    width: 1,
    height: 40,
    backgroundColor: colors.border,
  },
  statsCardPlaceholder: {
    marginHorizontal: spacing.lg,
    backgroundColor: colors.surface,
    borderRadius: borderRadius.lg,
    padding: spacing.lg,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.sm,
  },
  loadingText: {
    fontSize: fontSize.sm,
    color: colors.textSecondary,
  },

  // Quick Actions
  quickActions: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.lg,
  },
  actionCard: {
    alignItems: 'center',
    width: 100,
  },
  actionIcon: {
    width: 56,
    height: 56,
    borderRadius: borderRadius.lg,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: spacing.sm,
  },
  actionText: {
    fontSize: fontSize.sm,
    fontWeight: '600',
    color: colors.text,
  },

  // Sections
  section: {
    paddingHorizontal: spacing.lg,
    marginTop: spacing.md,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: spacing.md,
  },
  sectionTitle: {
    fontSize: fontSize.xl,
    fontWeight: '700',
    color: colors.text,
  },
  seeAll: {
    fontSize: fontSize.md,
    color: colors.primary,
    fontWeight: '600',
  },

  // Empty State
  emptyState: {
    alignItems: 'center',
    paddingVertical: spacing.xxl,
    backgroundColor: colors.surface,
    borderRadius: borderRadius.lg,
  },
  emptyText: {
    fontSize: fontSize.md,
    color: colors.textSecondary,
    marginTop: spacing.sm,
    marginBottom: spacing.md,
  },
  createButton: {
    backgroundColor: colors.primary,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.sm,
    borderRadius: borderRadius.sm,
  },
  createButtonText: {
    color: colors.textInverse,
    fontWeight: '600',
  },

  // Group Cards
  groupCard: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderRadius: borderRadius.md,
    padding: spacing.md,
    marginBottom: spacing.sm,
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.05,
    shadowRadius: 4,
  },
  groupAvatar: {
    width: 48,
    height: 48,
    borderRadius: borderRadius.full,
    backgroundColor: colors.primaryLight + '30',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: spacing.md,
  },
  groupAvatarText: {
    fontSize: fontSize.xl,
    fontWeight: '700',
    color: colors.primary,
  },
  groupInfo: {
    flex: 1,
  },
  groupName: {
    fontSize: fontSize.lg,
    fontWeight: '600',
    color: colors.text,
  },
  groupMembers: {
    fontSize: fontSize.sm,
    color: colors.textSecondary,
    marginTop: 2,
  },

  // Activity Cards
  activityCard: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderRadius: borderRadius.md,
    padding: spacing.md,
    marginBottom: spacing.sm,
    elevation: 1,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.03,
    shadowRadius: 3,
  },
  activityIconWrap: {
    width: 40,
    height: 40,
    borderRadius: borderRadius.full,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: spacing.md,
  },
  activityContent: {
    flex: 1,
  },
  activityTitle: {
    fontSize: fontSize.md,
    fontWeight: '600',
    color: colors.text,
  },
  activityDetail: {
    fontSize: fontSize.sm,
    color: colors.textSecondary,
    marginTop: 2,
  },
  activityMeta: {
    flexDirection: 'row',
    alignItems: 'center',
    marginTop: 4,
    gap: spacing.sm,
  },
  activityGroup: {
    fontSize: fontSize.xs,
    color: colors.primary,
    fontWeight: '600',
    backgroundColor: colors.primary + '10',
    paddingHorizontal: 6,
    paddingVertical: 2,
    borderRadius: 4,
    overflow: 'hidden',
  },
  activityTime: {
    fontSize: fontSize.xs,
    color: colors.textLight,
  },
  activityAmount: {
    fontSize: fontSize.md,
    fontWeight: '700',
    color: colors.primary,
    marginLeft: spacing.sm,
  },

  // Top Groups
  topGroupCard: {
    backgroundColor: colors.surface,
    borderRadius: borderRadius.md,
    padding: spacing.md,
    marginBottom: spacing.sm,
    elevation: 1,
    shadowColor: '#000',
    shadowOffset: {width: 0, height: 1},
    shadowOpacity: 0.03,
    shadowRadius: 3,
  },
  topGroupHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: spacing.sm,
  },
  topGroupRank: {
    width: 28,
    height: 28,
    borderRadius: borderRadius.full,
    backgroundColor: colors.primary + '15',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: spacing.sm,
  },
  topGroupRankText: {
    fontSize: fontSize.sm,
    fontWeight: '700',
    color: colors.primary,
  },
  topGroupInfo: {
    flex: 1,
  },
  topGroupName: {
    fontSize: fontSize.md,
    fontWeight: '600',
    color: colors.text,
  },
  topGroupBills: {
    fontSize: fontSize.xs,
    color: colors.textSecondary,
    marginTop: 1,
  },
  topGroupAmount: {
    fontSize: fontSize.md,
    fontWeight: '700',
    color: colors.text,
  },
  topGroupBarBg: {
    height: 6,
    backgroundColor: colors.borderLight,
    borderRadius: 3,
    overflow: 'hidden',
  },
  topGroupBarFill: {
    height: '100%',
    borderRadius: 3,
  },
});
